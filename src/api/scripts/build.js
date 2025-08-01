#!/usr/bin/env node
/**
 * Build script for @virtuoso/api-layer package
 * Prepares the module for npm publication
 */

const fs = require('fs');
const path = require('path');

console.log('Building @virtuoso/api-layer...');

// Ensure all directories exist
const dirs = [
  'abstractions',
  'config',
  'core',
  'utils',
  'examples'
];

dirs.forEach(dir => {
  const dirPath = path.join(__dirname, '..', dir);
  if (!fs.existsSync(dirPath)) {
    console.error(`Error: Required directory ${dir} does not exist`);
    process.exit(1);
  }
});

// Verify main entry point
const mainFile = path.join(__dirname, '..', 'index.js');
if (!fs.existsSync(mainFile)) {
  console.error('Error: Main entry point index.js does not exist');
  process.exit(1);
}

// Create LICENSE file if it doesn't exist
const licensePath = path.join(__dirname, '..', 'LICENSE');
if (!fs.existsSync(licensePath)) {
  const licenseContent = `MIT License

Copyright (c) ${new Date().getFullYear()} Virtuoso Team

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`;
  
  fs.writeFileSync(licensePath, licenseContent);
  console.log('Created LICENSE file');
}

// Verify package.json
const packageJsonPath = path.join(__dirname, '..', 'package.json');
try {
  const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
  
  // Check required fields
  const requiredFields = ['name', 'version', 'description', 'main', 'license'];
  const missingFields = requiredFields.filter(field => !packageJson[field]);
  
  if (missingFields.length > 0) {
    console.error(`Error: Missing required fields in package.json: ${missingFields.join(', ')}`);
    process.exit(1);
  }
  
  console.log(`Building ${packageJson.name} v${packageJson.version}`);
} catch (error) {
  console.error('Error reading package.json:', error.message);
  process.exit(1);
}

// Create .npmignore if it doesn't exist
const npmignorePath = path.join(__dirname, '..', '.npmignore');
if (!fs.existsSync(npmignorePath)) {
  const npmignoreContent = `# Development files
*.test.js
*.spec.js
__tests__/
coverage/
.eslintrc*
jest.config.js

# Build files
scripts/
*.log

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db`;
  
  fs.writeFileSync(npmignorePath, npmignoreContent);
  console.log('Created .npmignore file');
}

console.log('Build completed successfully!');
console.log('\nPackage contents:');
dirs.forEach(dir => {
  console.log(`  - ${dir}/`);
});
console.log('  - index.js');
console.log('  - package.json');
console.log('  - README.md');
console.log('  - LICENSE');