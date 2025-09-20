FROM node:19-alpine
RUN apk update && apk add xdg-utils
WORKDIR /markdown-frontend
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

ARG FRONTEND_HOST
ARG FRONTEND_PORT
ENV HOST=${FRONTEND_HOST}
ENV PORT=${FRONTEND_PORT}
EXPOSE ${PORT}

CMD ["npm", "run", "dev"]
