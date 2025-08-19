# Docker Swarm Secrets

This directory contains example secret files for Docker Swarm deployment.

## Setup Instructions

1. Copy the example files and remove the `.example` extension:
   ```bash
   cp postgres_password.txt.example postgres_password.txt
   cp jwt_secret.txt.example jwt_secret.txt
   cp redis_password.txt.example redis_password.txt
   cp typesense_api_key.txt.example typesense_api_key.txt
   cp minio_root_password.txt.example minio_root_password.txt
   ```

2. Edit each file with your actual secure passwords and keys.

3. Ensure proper file permissions:
   ```bash
   chmod 600 *.txt
   ```

## Security Notes

- Never commit actual secret files to version control
- Use strong, randomly generated passwords
- Rotate secrets regularly
- The actual `.txt` files are already in `.gitignore`

## Docker Swarm Secret Creation

The deployment script will automatically create Docker Swarm secrets from these files.
Alternatively, you can create them manually:

```bash
docker secret create postgres_password postgres_password.txt
docker secret create jwt_secret jwt_secret.txt
docker secret create redis_password redis_password.txt
docker secret create typesense_api_key typesense_api_key.txt
docker secret create minio_root_password minio_root_password.txt
```