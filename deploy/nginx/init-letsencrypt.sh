echo "### Запускаем Nginx..."
docker compose up -d nginx

# УДАЛИТЕ ИЛИ ЗАКОММЕНТИРУЙТЕ БЛОК С "rm -rf"
# echo "### Удаляем временный сертификат..."
# docker compose run --rm --entrypoint \
#   "rm -rf /etc/letsencrypt/live/$DOMAIN /etc/letsencrypt/archive/$DOMAIN /etc/letsencrypt/renewal/$DOMAIN.conf" certbot

echo "### Запрашиваем настоящий сертификат у Let's Encrypt..."
staging_arg=""
if [ $STAGING != 0 ]; then staging_arg="--staging"; fi

# Добавляем флаг --keep-until-expiring или удаляем --force-renewal,
# чтобы certbot просто обновил существующую заглушку
docker compose run --rm --entrypoint \
  "certbot certonly --webroot -w /var/www/certbot \
    $staging_arg \
    -d $DOMAIN \
    --email $EMAIL \
    --rsa-key-size 4096 \
    --agree-tos \
    --force-renewal \
    --non-interactive" certbot

echo "### Перезапускаем Nginx для применения настоящих сертификатов..."
docker compose exec nginx nginx -s reload
