export JWT_SECRET=chatsecret
export MONGODB_URI=mongodb://localhost:27017/chat_app
run:
	go run main.go

certbot:
	docker-compose run --rm --entrypoint "\
  certbot certonly --webroot -w /var/www/certbot \
    -d chat-api.odds.team \
    --register-unsafely-without-email --agree-tos --force-renewal" certbot