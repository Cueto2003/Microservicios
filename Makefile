test-auth-service-register:
	curl -X POST http://localhost:8082/Auth-Server/register \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "email=oscar@example.com" \
        -d "password=SuperSecreta123\!" \
        -d "role=user" \
        -d "provider=local" \
        -d "full_name=Oscar Cueto" \
        -d "avatar_url=https://ejemplo.com/avatar.png" \
        -d "phone_number=123456789" \
        -d "birth_date=1995-06-21"

test-metadataUser:
	curl -X GET "http://localhost:8081/MetadataUser/Get?email=oscar@example.com"

run-consul-dev-server:
	docker run -d --name consul \
		--network microservice-net \
		-p 8500:8500 -p 8600:8600/udp \
		consul:1.15.4 agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0