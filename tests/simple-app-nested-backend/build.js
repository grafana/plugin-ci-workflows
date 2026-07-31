// Copies the static plugin files to dist/. This fixture has no webpack build:
// it exists to test the packaging of an app with its own backend AND a nested
// datasource backend, so the frontend is intentionally just static files.
// The backend binaries are built separately by `mage buildAll` (see Magefile.go).
const fs = require('fs');
const path = require('path');

const files = [
  ['src/plugin.json', 'dist/plugin.json'],
  ['src/module.js', 'dist/module.js'],
  ['src/module.js.map', 'dist/module.js.map'],
  ['src/img/logo.svg', 'dist/img/logo.svg'],
  ['src/datasource/plugin.json', 'dist/datasource/plugin.json'],
  ['src/datasource/module.js', 'dist/datasource/module.js'],
  ['CHANGELOG.md', 'dist/CHANGELOG.md'],
  ['LICENSE', 'dist/LICENSE'],
  ['README.md', 'dist/README.md'],
];

fs.rmSync('dist', { recursive: true, force: true });
for (const [src, dst] of files) {
  fs.mkdirSync(path.dirname(dst), { recursive: true });
  fs.copyFileSync(src, dst);
}
