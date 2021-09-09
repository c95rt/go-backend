## Create the env file

Create a file named `dev.env` with the env variables in the root of the project

## Run the server
1. Install Docker and Docker Compose (Linux)
2. In the root of the project, run the command: ```docker-compose up --build```

## Another way to run the server
1. Install Go
2. Go to the root of the project
3. Run the command: ```go run . backend-up``` or ```PORT=:port go run . backend-up``` (changing :port with the desired port)

## See the API documentation

Postman link:
https://www.getpostman.com/collections/4645cdaa3952105320e2
(Remember to change the env var `{{parque_oasis_backend_local}}` with `localhost:3001`)