const fs = require('fs');
const path = require('path');

async function buildPrComment() {
  // Ensure we are in the right directory
  const reportsDir = 'all-reports';
  if (!fs.existsSync(reportsDir) || !fs.statSync(reportsDir).isDirectory()) {
    console.error('Failed to enter directory all-reports');
    process.exit(1);
  }

  // Initialize the table variable
  let table = '### Playwright test results';

  // Check if any summary.txt has a PLUGIN_NAME value
  const summaryFiles = fs
    .readdirSync(reportsDir)
    .map((dir) => path.join(reportsDir, dir, 'summary.txt'))
    .filter((file) => fs.existsSync(file));

  const usePluginName = summaryFiles.some((file) => {
    const content = fs.readFileSync(file, 'utf8');
    const match = content.match(/^PLUGIN_NAME=(.+)$/m);
    return match && match[1].trim() !== '';
  });

  if (usePluginName) {
    table +=
      '\n| Plugin Name | Image Name | Version | Result | Report |\n|:----------- |:---------- |:------- |:------: |:------: |';
  } else {
    table += '\n| Image Name | Version | Result | Report |\n|:---------- |:------- |:------: |:------: |';
  }

  // Initialize an array to store rows
  let rows = [];
  let uploadReportDisabled = false;

  const reportBaseUrl = process.env.REPORT_BASE_URL;
  if (!reportBaseUrl) {
    console.error('REPORT_BASE_URL environment variable is required but was not set.');
    process.exit(1);
  }

  // Iterate through subdirectories
  fs.readdirSync(reportsDir).forEach((dir) => {
    const dirPath = path.join(reportsDir, dir);
    if (!fs.statSync(dirPath).isDirectory()) return;

    const summaryFile = path.join(dirPath, 'summary.txt');
    if (!fs.existsSync(summaryFile)) {
      console.warn(`Warning: summary.txt not found in ${dir}`);
      return;
    }

    // Read data from summary.txt
    const content = fs.readFileSync(summaryFile, 'utf8');
    const getValue = (key) => (content.match(new RegExp(`${key}=(.*)`)) || [])[1]?.trim() || '';

    const grafanaImage = getValue('GRAFANA_IMAGE');
    const grafanaVersion = getValue('GRAFANA_VERSION');
    const testOutput = getValue('OUTPUT');
    const pluginName = getValue('PLUGIN_NAME');
    uploadReportDisabled = getValue('UPLOAD_REPORT_ENABLED') === 'false';

    // Construct report link
    const reportLink = `${reportBaseUrl}/${dir}/index.html`;

    // Map result to emoji
    const resultEmoji = testOutput === 'success' ? '✅' : '❌';

    // Check for index.html
    const hasReport = fs.existsSync(path.join(dirPath, 'index.html'));
    const reportCell = hasReport ? `[View report](${reportLink})` : ' ';

    // Add row to table
    if (usePluginName) {
      rows.push(`| ${pluginName} | ${grafanaImage} | ${grafanaVersion} | ${resultEmoji} | ${reportCell} |`);
    } else {
      rows.push(`| ${grafanaImage} | ${grafanaVersion} | ${resultEmoji} | ${reportCell} |`);
    }
  });

  // Sort rows by version (assuming <major>.<minor>.<patch> format)
  rows.sort((a, b) => {
    const getVersion = (row) => row.split('|')[usePluginName ? 3 : 2].trim();
    return getVersion(b).localeCompare(getVersion(a), undefined, { numeric: true });
  });

  // Add sorted rows to table
  table += '\n' + rows.join('\n') + '\n';

  const ciLink = `https://github.com/${process.env.GITHUB_REPOSITORY_OWNER}/${process.env.GITHUB_REPOSITORY_NAME}/blob/${process.env.DEFAULT_BRANCH}/.github/workflows/ci.yml`;
  if (uploadReportDisabled) {
    table += `\n⚠️  To make Playwright reports for failed tests accessible, set the \`upload-report\` input to \`true\` in your [CI workflow](${ciLink}). For more details, refer to the [Developer Portal documentation](https://grafana.com/developers/plugin-tools/e2e-test-a-plugin/ci).\n`;
  }

  table += `\n> ℹ️ Reports require a Grafana Google Workspace sign-in to view and are retained for 90 days.`;

  console.log(table);
}

buildPrComment();
