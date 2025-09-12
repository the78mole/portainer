FROM php:8.2-apache

# Copy HTML files to the web root
COPY html/ /var/www/html/

# Set proper permissions
RUN chown -R www-data:www-data /var/www/html/ && \
    chmod -R 644 /var/www/html/

# Enable Apache modules for better PHP support
RUN a2enmod rewrite

# Set the default directory index
RUN echo "DirectoryIndex index.php index.html" >> /etc/apache2/apache2.conf

EXPOSE 80