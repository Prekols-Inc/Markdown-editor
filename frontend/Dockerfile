FROM node:19-alpine
RUN apk update && apk add xdg-utils
WORKDIR /markdown-frontend
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

COPY entrypoint.sh ./
RUN chmod +x entrypoint.sh

EXPOSE ${FRONTEND_PORT}

CMD ./entrypoint.sh