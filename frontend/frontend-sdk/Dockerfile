FROM node:16.16-alpine
WORKDIR /app
ENV PATH /app/node_modules/.bin $PATH

COPY package.json /app
COPY package-lock.json /app

RUN npm ci --silent /app
COPY . /app
RUN npm run build

FROM nginx:stable-alpine
COPY --from=build /app/dist/sdk.* /usr/share/nginx/html

COPY nginx/default.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
