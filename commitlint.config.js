module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [
      2,
      'always',
      [
        'feat',     // New feature (official)
        'fix',      // Bug fix (official)
        'docs',     // Documentation changes
        'style',    // Code style changes
        'refactor', // Code refactoring
        'perf',     // Performance improvements
        'test',     // Test changes
        'build',    // Build system changes
        'ci',       // CI configuration changes
        'chore'     // Maintenance tasks
      ]
    ],
    'type-case': [2, 'always', 'lower-case'],
    'type-empty': [2, 'never'],
    'subject-empty': [2, 'never'],
    'subject-full-stop': [2, 'never', '.'],
    'subject-case': [0],
    'scope-case': [2, 'always', 'lower-case']
  }
};
