FROM bincooo/chrome-vnc:latest

ADD . /app
WORKDIR /app

RUN npm install
ENTRYPOINT ["tail","-f","/dev/null"]
#ENTRYPOINT ["/bin/bash", "/app/docker-entrypoint.sh"]
# Xvfb :99 -ac & export DISPLAY=:99
# x11vnc -display :99 -forever -bg -o /var/log/x11vnc.log -rfbport 5900