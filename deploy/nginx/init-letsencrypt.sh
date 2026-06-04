#!/bin/bash

# --- НАСТРОЙКИ ---
DOMAIN="b2b-courier-14.ru" # Замените на ваш домен
EMAIL="rusneustroevkz@gmail.com"
STAGING=0 # Поставьте 1 для теста, чтобы не поймать лимиты Let's Encrypt

# Пути к папкам
DATA_PATH="./deploy/nginx/certbot"
CONF_PATH="./deploy/nginx/conf.d"

if [ -d "$DATA_PATH" ]; then
  read -p "Папка с сертификатами уже существует. Удалить и начать заново? (y/N) " decision
  if [ "$decision" != "Y" ] && [ "$decision" != "y" ]; then
    exit
  fi
fi

echo "### Создаем временный сертификат для $DOMAIN..."
mkdir -p "$DATA_PATH/conf/live/$DOMAIN"
openssl req -x509 -nodes -newkey rsa:2048 -days 1 \
  -keyout "$DATA_PATH/conf/live/$DOMAIN/privkey.pem" \
  -out "$DATA_PATH/conf/live/$DOMAIN/fullchain.pem" \
  -subj "/CN=localhost"

echo "### Запускаем Nginx..."
docker compose up -d nginx

echo "### Удаляем временный сертификат..."
docker compose run --rm --entrypoint \
  "rm -rf /etc/letsencrypt/live/$DOMAIN /etc/letsencrypt/archive/$DOMAIN /etc/letsencrypt/renewal/$DOMAIN.conf" certbot

echo "### Запрашиваем настоящий сертификат у Let's Encrypt..."
# Выбираем режим (тестовый или боевой)
staging_arg=""
if [ $STAGING != 0 ]; then staging_arg="--staging"; fi

docker compose run --rm --entrypoint \
  "certbot certonly --webroot -w /var/www/certbot \
    $staging_arg \
    -d $DOMAIN \
    --email $EMAIL \
    --rsa-key-size 4096 \
    --agree-tos \
    --force-renewal \
    --non-interactive" certbot

echo "### Перезапускаем Nginx для применения сертификатов..."
docker compose exec nginx nginx -s reload
