# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability within this provider, please report it responsibly.

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to the maintainers or use GitHub's private vulnerability reporting feature.

### What to Include

- Type of vulnerability
- Full paths of source file(s) related to the vulnerability
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the vulnerability

### Response Timeline

- We will acknowledge receipt of your vulnerability report within 48 hours
- We will provide a more detailed response within 7 days
- We will work to fix the vulnerability and release a patch as soon as possible

## Security Best Practices for Users

1. **Protect your API tokens**: Never commit API tokens to version control. Use environment variables or secure secret management.

2. **Use environment variables**:
   ```shell
   export STATUSGATOR_API_TOKEN="your-token"
   ```

3. **Terraform state**: Your Terraform state may contain sensitive information. Always use encrypted remote state storage.

4. **Access control**: Limit who has access to your Terraform configurations and state files.
