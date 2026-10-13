// Runs the plopfile generator through node-plop, without the plop CLI.
// Writes the provisioning files and docker-compose.e2e.yaml into the current working directory.
import nodePlop from "node-plop";
import path from "path";
import { fileURLToPath } from "url";

const plopfile = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  "plopfile.mjs",
);

const plop = await nodePlop(plopfile, { destBasePath: process.cwd() });
const { changes, failures } = await plop
  .getGenerator("e2e-testing-provisioning")
  .runActions({});

for (const change of changes) {
  console.log(`✔ ${change.type} ${change.path}`);
}
for (const failure of failures) {
  console.error(`✖ ${failure.type} ${failure.path} ${failure.error}`);
}
if (failures.length > 0) {
  process.exit(1);
}
