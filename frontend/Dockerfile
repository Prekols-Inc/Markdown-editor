FROM node:19-alpine
RUN apk update && apk add xdg-utils
WORKDIR /markdown-frontend
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

EXPOSE ${FRONTEND_PORT}

CMD sh -c "npm test && npm run release || exit 1"