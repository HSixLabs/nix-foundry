# Contributing

## Development Setup

1. Fork and clone the repository:

   ```bash
   git clone https://github.com/yourusername/nix-foundry.git
   cd nix-foundry
   ```

2. Install dependencies:

   ```bash
   go mod download
   ```

3. Install pre-commit hooks:
   ```bash
   pre-commit install
   ```

## Code Style

- Follow Go best practices and idioms
- Add docstrings to packages and functions
- Keep functions focused and concise
- Write tests for new functionality

## Pull Request Process

1. Create a feature branch:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and commit:

   ```bash
   git add .
   git commit -m "feat: add your feature"
   ```

3. Push to your fork:

   ```bash
   git push origin feature/your-feature-name
   ```

4. Open a Pull Request with:
   - Clear description of changes
   - Any related issues
   - Test coverage
   - Documentation updates

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `style:` Code style changes
- `refactor:` Code refactoring
- `test:` Adding or modifying tests
- `chore:` Maintenance tasks

## Testing

- Write unit tests for new functionality
- Ensure all tests pass:
  ```bash
  go test ./...
  ```

## Documentation

- Update documentation for new features
- Keep README.md current
- Add docstrings to new functions
- Run doc generation:
  ```bash
  go run scripts/gendocs.go
  ```
