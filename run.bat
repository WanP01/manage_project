chcp 65001
cd project-user
docker build -t project-user:latest .
cd ..
cd project-api
docker build -t project-api:latest .
cd ..
cd project-project
docker build -t project-project:latest .
cd ..
docker-compose up -d