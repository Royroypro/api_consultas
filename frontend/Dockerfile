FROM nginx:1.25-alpine

# Copiar el archivo estático index.html directamente al directorio de Nginx
COPY index.html /usr/share/nginx/html/index.html

# Exponer puerto 80
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
