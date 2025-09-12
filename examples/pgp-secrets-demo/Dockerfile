FROM php:8.2-apache

# Copy HTML files to the web root
COPY html/ /var/www/html/

# Set proper permissions for Apache
RUN chown -R www-data:www-data /var/www/html/ && \
    chmod -R 755 /var/www/html/ && \
    find /var/www/html/ -type f -exec chmod 644 {} \;

# Enable Apache modules for better PHP support
RUN a2enmod rewrite

# Set the default directory index
RUN echo "DirectoryIndex index.php index.html" >> /etc/apache2/apache2.conf

# Disable .htaccess processing to avoid permission issues
RUN echo "AllowOverride None" >> /etc/apache2/apache2.conf

EXPOSE 80