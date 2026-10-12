module.exports = async ({ core, exec }) => {
  const fs = require("fs");
  const path = require("path");

  const mode = (process.env.SCRIPTS_MODE || "").trim();
  const agent = process.env.PM_AGENT;
  const frozen = (process.env.FROZEN_INSTALL_CMD || "").trim();
  const pluginDir = process.env.PLUGIN_DIRECTORY || ".";

  // Compares major.minor.patch, ignoring any prerelease/build suffix.
  const parse = (v) => v.trim().replace(/^v/, "").split(/[-+]/)[0].split(".").map((n) => parseInt(n, 10) || 0);
  const cmp = (a, b) => {
    const [x, y] = [parse(a), parse(b)];
    for (let i = 0; i < 3; i++) {
      if ((x[i] || 0) !== (y[i] || 0)) return (x[i] || 0) < (y[i] || 0) ? -1 : 1;
    }
    return 0;
  };
  const atLeast = (v, min) => cmp(v, min) >= 0;

  // Runtime version of the binary that will actually run the install (not the declared one).
  let runtime = "";
  const runtimeVersion = async (bin) => {
    const out = await exec.getExecOutput(bin, ["--version"], { cwd: pluginDir, silent: true });
    runtime = out.stdout.trim();
    return runtime;
  };

  // A missing or unparseable package.json counts as "no allowlist configuration".
  const readPackageJson = () => {
    try {
      return JSON.parse(fs.readFileSync(path.join(pluginDir, "package.json"), "utf8"));
    } catch (e) {
      core.info(`Could not read package.json: ${e.message}`);
      return {};
    }
  };

  const isPlainObject = (v) => v !== null && typeof v === "object" && !Array.isArray(v);

  if (!["none", "allowlist", "all"].includes(mode)) {
    core.setFailed(`Invalid node-package-manager-scripts value "${mode}": must be one of none, allowlist, all`);
    return;
  }
  if (!frozen) {
    core.setFailed("Could not determine the frozen install command of the package manager");
    return;
  }

  let cmd = frozen;
  // Yarn Berry (v2+) only: value for YARN_ENABLE_SCRIPTS
  // (env form of `enableScripts` in .yarnrc.yml). Stays empty for
  // npm, pnpm and Yarn Classic, which ignores it: its command uses
  // --ignore-scripts for "none" and the "allowlist" fallback.
  let yarnEnable = "";
  let effective = mode;
  let fallbackReason = "";

  try {
    const known = ["npm", "pnpm", "yarn", "yarnBerry"].includes(agent);
    if (!known) {
      if (mode !== "all") {
        core.setFailed(`Unknown package manager agent "${agent}": cannot apply install scripts policy "${mode}"`);
        return;
      }
    } else if (mode === "none") {
      if (agent === "yarnBerry") {
        const major = parse(await runtimeVersion("yarn"))[0];
        cmd = `${frozen} ${major === 2 ? "--skip-builds" : "--mode=skip-build"}`;
        yarnEnable = "false";
      } else {
        cmd = `${frozen} --ignore-scripts`;
      }
    } else if (mode === "allowlist") {
      if (agent === "yarnBerry") {
        yarnEnable = "false";
      } else if (agent === "npm") {
        const v = await runtimeVersion("npm");
        const allowScripts = readPackageJson().allowScripts;
        if (atLeast(v, "12.0.0")) {
          // npm 12 blocks unlisted dependency scripts natively.
        } else if (atLeast(v, "11.16.0") && isPlainObject(allowScripts)) {
          // npm 11 still runs unlisted scripts: strict mode fails the install instead.
          cmd = `${frozen} --strict-allow-scripts`;
        } else if (!atLeast(v, "11.16.0")) {
          fallbackReason = `npm ${v} does not support allowScripts (requires npm 11.16.0 or later)`;
        } else {
          fallbackReason = 'package.json has no "allowScripts" field';
        }
      } else if (agent === "pnpm") {
        const v = await runtimeVersion("pnpm");
        const pnpmCfg = readPackageJson().pnpm || {};
        if (parse(v)[0] >= 10) {
          // pnpm 10+ blocks unlisted dependency builds natively.
        } else if (Array.isArray(pnpmCfg.onlyBuiltDependencies) || pnpmCfg.onlyBuiltDependenciesFile) {
          // pnpm < 10 only builds the listed dependencies.
        } else {
          fallbackReason = `pnpm ${v} has no build allowlist configured (pnpm.onlyBuiltDependencies)`;
        }
      } else {
        fallbackReason = "Yarn Classic does not support script allowlists";
      }
      if (fallbackReason) {
        effective = "none";
        cmd = `${frozen} --ignore-scripts`;
        core.warning(
          `All install scripts were disabled because ${fallbackReason}. ` +
            'Configure an install scripts allowlist for your package manager (see the "node-package-manager-scripts" input), ' +
            'or set "node-package-manager-scripts: all" to run every install script.',
          { title: "Install scripts disabled" },
        );
      }
    } else if (agent === "yarnBerry") {
      yarnEnable = "true";
    } else if (agent === "npm") {
      if (atLeast(await runtimeVersion("npm"), "11.16.0")) cmd = `${frozen} --dangerously-allow-all-scripts`;
    } else if (agent === "pnpm") {
      const v = await runtimeVersion("pnpm");
      if (atLeast(v, "10.9.0")) {
        cmd = `${frozen} --config.dangerously-allow-all-builds=true`;
      } else if (atLeast(v, "10.0.0")) {
        // pnpm 10.0-10.8 has no way to enable all builds: only onlyBuiltDependencies packages build.
        effective = "allowlist";
        core.warning(
          `pnpm ${v} cannot enable all dependency builds from CI (dangerouslyAllowAllBuilds requires pnpm 10.9.0 or later), ` +
            'so only packages listed in "onlyBuiltDependencies" will run install scripts. ' +
            'Upgrade pnpm to 10.9 or later, or list the needed packages in "onlyBuiltDependencies".',
          { title: "Not all install scripts enabled" },
        );
      }
    }
  } catch (e) {
    core.setFailed(`Could not determine the package manager runtime version: ${e.message}`);
    return;
  }

  core.info(`Install scripts policy: requested=${mode} effective=${effective} agent=${agent}`);
  core.info(`Install command: ${cmd}`);
  core.setOutput("install-cmd", cmd);
  core.setOutput("yarn-enable-scripts", yarnEnable);
  if (runtime) console.log(`::act-debug::msg="pm-runtime-version" value=${runtime}`);
  console.log(`::act-debug::msg="scripts-policy" requested=${mode} effective=${effective}`);
  console.log(`::act-debug::msg="install-cmd" cmd=${cmd}`);
  console.log(`::act-debug::msg="yarn-enable-scripts" value=${yarnEnable || "not-set"}`);
};
