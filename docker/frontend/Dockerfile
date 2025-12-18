FROM node:19-alpine
RUN apk update && apk add xdg-utils
WORKDIR /markdown-frontend
COPY frontend/package*.json .
RUN npm install
COPY ./frontend .
RUN npm run build

EXPOSE ${FRONTEND_PORT}

CMD sh -c "npm test && npm run release || exit 1"