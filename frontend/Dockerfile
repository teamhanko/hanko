FROM node:16.16-alpine as build

WORKDIR /app
ENV PATH /app/node_modules/.bin:$PATH

# frontend-sdk
COPY frontend-sdk/package.json frontend-sdk/
COPY frontend-sdk/package-lock.json frontend-sdk/

WORKDIR /app/frontend-sdk
RUN npm ci --silent
COPY frontend-sdk/ ./
RUN npm run build


# elements
WORKDIR /app/elements
COPY elements/package.json ./
COPY elements/package-lock.json ./

RUN npm ci
COPY ./elements ./
RUN npm run build

FROM nginx:stable-alpine
COPY --from=build /app/elements/dist/element.hanko-auth.js /usr/share/nginx/html
COPY --from=build /app/frontend-sdk/dist/sdk.* /usr/share/nginx/html

COPY elements/nginx/default.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
