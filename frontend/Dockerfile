FROM node:19-alpine
RUN apk update && apk add xdg-utils
WORKDIR /markdown-frontend
COPY package*.json ./
RUN npm install
COPY . .

ARG VITE_STORAGE_API_BASE_URL
ARG VITE_AUTH_API_BASE_URL
ENV VITE_STORAGE_API_BASE_URL=${VITE_STORAGE_API_BASE_URL}
ENV VITE_AUTH_API_BASE_URL=${VITE_AUTH_API_BASE_URL}

RUN npm run build

EXPOSE ${FRONTEND_PORT}

CMD ["npm", "run", "release"]
