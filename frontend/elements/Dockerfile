FROM node:16.16-alpine as build

WORKDIR /app
ENV PATH /app/node_modules/.bin:$PATH

# Build SDK dependency
COPY ../frontend-sdk/package.json ../frontend-sdk/
COPY ../frontend-sdk/package-lock.json ../frontend-sdk/

RUN npm ci --silent ../frontend-sdk
COPY ../frontend-sdk/ ../frontend-sdk
RUN npm run build ../frontend-sdk

COPY elements/package.json ./
COPY elements/package-lock.json ./

RUN npm ci --silent
COPY . ./
RUN npm run build

FROM nginx:stable-alpine
COPY --from=build /app/elements/dist/elements.js /usr/share/nginx/html

COPY elements/nginx/default.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
