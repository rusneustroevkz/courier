echo "### Запрашиваем настоящий сертификат у Let's Encrypt..."
staging_arg=""
if [ $STAGING != 0 ]; then staging_arg="--staging"; fi

docker compose run \
  --rm certbot certonly \
  --webroot \
  --webroot-path=/var/www/certbot \
  -d $DOMAIN \
  --email $EMAIL \
  --agree-tos \
  --no-eff-email

echo "### Перезапускаем Nginx для применения сертификатов..."
docker compose exec nginx nginx -s reload
