# Simple static server for plain HTML/JS/CSS with Nginx
FROM nginx:alpine

# Nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Static assets
COPY . /usr/share/nginx/html

EXPOSE 80