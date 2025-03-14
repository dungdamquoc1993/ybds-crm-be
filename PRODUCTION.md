# YBDS Production Deployment Guide

This guide provides instructions for deploying the YBDS application in a production environment, with a focus on handling image storage.

## Prerequisites

- Docker and Docker Compose installed on the server
- Git installed on the server
- A domain name (optional, but recommended)
- SSL certificate (optional, but recommended)

## Environment Variables

Create a `.env.prod` file in the root directory with the following variables:

```
DB_USER=your_db_user
DB_PASS=your_secure_password
DB_NAME=ybds
JWT_SECRET=your_secure_jwt_secret
```

Make sure to use strong, secure passwords and secrets.

## Deployment Steps

1. Clone the repository:

```bash
git clone https://github.com/yourusername/ybds.git
cd ybds
```

2. Create the necessary directories:

```bash
mkdir -p backups
```

3. Start the application using the production Docker Compose file:

```bash
docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
```

4. Verify that the application is running:

```bash
docker-compose -f docker-compose.prod.yml ps
```

## Image Storage

In the production environment, images are stored in a Docker volume named `uploads-data`. This volume is mounted to the `/app/uploads` directory in the container.

### Backup Strategy

The production Docker Compose file includes a backup service that creates daily backups of the uploads directory. The backups are stored in the `./backups` directory on the host machine and are kept for 7 days.

To manually trigger a backup:

```bash
docker-compose -f docker-compose.prod.yml exec backup sh -c "tar -czf /backups/uploads-backup-manual-$(date +%Y%m%d-%H%M%S).tar.gz -C /data uploads"
```

### Restoring from Backup

To restore from a backup:

1. Stop the application:

```bash
docker-compose -f docker-compose.prod.yml down
```

2. Extract the backup to a temporary directory:

```bash
mkdir -p temp
tar -xzf backups/uploads-backup-YYYYMMDD-HHMMSS.tar.gz -C temp
```

3. Start the application:

```bash
docker-compose -f docker-compose.prod.yml up -d
```

4. Copy the extracted files to the container:

```bash
docker cp temp/uploads/. ybds-app:/app/uploads/
```

5. Clean up:

```bash
rm -rf temp
```

## Scaling Considerations

For high-traffic applications, consider the following:

1. **External Storage**: Instead of using a Docker volume, consider using an external storage solution like AWS S3, Google Cloud Storage, or Azure Blob Storage.

2. **CDN**: Use a Content Delivery Network (CDN) to serve images, reducing the load on your application server.

3. **Load Balancing**: Deploy multiple instances of the application behind a load balancer.

## Monitoring

Monitor the disk usage of the uploads directory to ensure you don't run out of space:

```bash
docker exec ybds-app df -h /app/uploads
```

## Troubleshooting

If images are not being served correctly:

1. Check that the uploads directory exists and has the correct permissions:

```bash
docker exec ybds-app ls -la /app/uploads
```

2. Verify that the UPLOAD_DIR environment variable is set correctly:

```bash
docker exec ybds-app env | grep UPLOAD_DIR
```

3. Check the application logs for any errors:

```bash
docker-compose -f docker-compose.prod.yml logs app
``` 