echo "### Запрашиваем настоящий сертификат у Let's Encrypt..."
staging_arg=""
if [ $STAGING != 0 ]; then staging_arg="--staging"; fi

# МЕНЯЕМ "run --rm --entrypoint" НА "exec"
docker compose exec certbot certbot certonly --webroot -w /var/www/certbot \
    $staging_arg \
    -d $DOMAIN \
    --email $EMAIL \
    --rsa-key-size 4096 \
    --agree-tos \
    --force-renewal \
    --non-interactive

echo "### Перезапускаем Nginx для применения сертификатов..."
docker compose exec nginx nginx -s reload
