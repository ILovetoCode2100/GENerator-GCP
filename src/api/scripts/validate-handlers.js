#!/usr/bin/env node
/**
 * Validation script for API handlers
 * Ensures all handlers are properly implemented
 */

const fs = require('fs');
const path = require('path');

console.log('Validating API handlers...');

// Expected handlers and their required methods
const expectedHandlers = {
  'project-handler.js': ['createProject', 'listProjects', 'listProjectGoals'],
  'goal-handler.js': ['createGoal', 'getGoalVersions', 'executeGoalSnapshot'],
  'journey-handler.js': ['createJourney', 'createJourneyAlt', 'listJourneysWithStatus', 'getJourneyDetails', 'updateJourney', 'attachCheckpoint', 'attachLibraryCheckpoint'],
  'checkpoint-handler.js': ['createCheckpoint', 'createCheckpointAlt', 'getCheckpointDetails', 'getCheckpointSteps', 'addCheckpointToLibrary'],
  'step-handler.js': ['addTestStep', 'addTestStepAlt', 'getStepDetails', 'updateStepProperties'],
  'execution-handler.js': ['executeGoal', 'getExecutionStatus', 'getExecutionAnalysis'],
  'library-handler.js': ['addToLibrary', 'getLibraryCheckpoint', 'updateLibraryCheckpoint', 'removeLibraryStep', 'moveLibraryStep'],
  'data-handler.js': ['createDataTable', 'getDataTable', 'importDataToTable'],
  'environment-handler.js': ['createEnvironment']
};

let hasErrors = false;

// Check each handler file
Object.entries(expectedHandlers).forEach(([filename, methods]) => {
  const filePath = path.join(__dirname, '..', 'core', 'handlers', filename);
  
  console.log(`\nChecking ${filename}...`);
  
  if (!fs.existsSync(filePath)) {
    console.error(`  ❌ File not found: ${filename}`);
    hasErrors = true;
    return;
  }
  
  try {
    const fileContent = fs.readFileSync(filePath, 'utf8');
    
    // Check if handler class is exported
    const handlerName = filename.replace('.js', '').split('-').map(w => 
      w.charAt(0).toUpperCase() + w.slice(1)
    ).join('');
    
    if (!fileContent.includes(`class ${handlerName}`)) {
      console.error(`  ❌ Handler class ${handlerName} not found`);
      hasErrors = true;
    }
    
    // Check for required methods
    methods.forEach(method => {
      // Look for method definition in various formats
      const patterns = [
        `async ${method}(`,
        `${method}(`,
        `${method} (`,
        `async ${method} (`
      ];
      
      const hasMethod = patterns.some(pattern => fileContent.includes(pattern));
      
      if (!hasMethod) {
        console.error(`  ❌ Method not found: ${method}`);
        hasErrors = true;
      } else {
        console.log(`  ✓ ${method}`);
      }
    });
    
    // Check if handler extends BaseHandler
    if (!fileContent.includes('extends BaseHandler')) {
      console.error(`  ❌ Handler does not extend BaseHandler`);
      hasErrors = true;
    }
    
  } catch (error) {
    console.error(`  ❌ Error reading file: ${error.message}`);
    hasErrors = true;
  }
});

// Check base handler exists
const baseHandlerPath = path.join(__dirname, '..', 'core', 'handlers', 'base-handler.js');
if (!fs.existsSync(baseHandlerPath)) {
  console.error('\n❌ BaseHandler not found');
  hasErrors = true;
} else {
  console.log('\n✓ BaseHandler exists');
}

// Check index.js exports all handlers
const indexPath = path.join(__dirname, '..', 'core', 'handlers', 'index.js');
if (fs.existsSync(indexPath)) {
  const indexContent = fs.readFileSync(indexPath, 'utf8');
  const handlerNames = Object.keys(expectedHandlers).map(f => 
    f.replace('.js', '').split('-').map(w => 
      w.charAt(0).toUpperCase() + w.slice(1)
    ).join('') + 'Handler'
  );
  
  console.log('\nChecking handler exports...');
  handlerNames.forEach(name => {
    if (!indexContent.includes(name)) {
      console.error(`  ❌ ${name} not exported`);
      hasErrors = true;
    } else {
      console.log(`  ✓ ${name} exported`);
    }
  });
} else {
  console.error('\n❌ Handler index.js not found');
  hasErrors = true;
}

// Check main API service
const mainIndexPath = path.join(__dirname, '..', 'index.js');
if (fs.existsSync(mainIndexPath)) {
  console.log('\n✓ Main index.js exists');
  
  const mainContent = fs.readFileSync(mainIndexPath, 'utf8');
  if (!mainContent.includes('class VirtuosoApiService')) {
    console.error('  ❌ VirtuosoApiService class not found');
    hasErrors = true;
  }
  if (!mainContent.includes('createApiService')) {
    console.error('  ❌ createApiService function not found');
    hasErrors = true;
  }
} else {
  console.error('\n❌ Main index.js not found');
  hasErrors = true;
}

// Summary
console.log('\n' + '='.repeat(50));
if (hasErrors) {
  console.error('❌ Validation failed! Please fix the errors above.');
  process.exit(1);
} else {
  console.log('✅ All handlers validated successfully!');
  console.log('\nSummary:');
  console.log(`  - ${Object.keys(expectedHandlers).length} handler files validated`);
  console.log(`  - ${Object.values(expectedHandlers).flat().length} methods checked`);
  console.log('  - All exports verified');
}