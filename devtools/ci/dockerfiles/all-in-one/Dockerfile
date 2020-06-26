FROM postgres:11-alpine
COPY goalert init.sql start.sh /bin/
EXPOSE 8081
CMD ["/bin/start.sh"]
