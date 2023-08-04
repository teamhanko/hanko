# pull official base image
FROM node:16-alpine

# set working directory
WORKDIR /app

# add `/app/node_modules/.bin` to $PATH
ENV PATH /app/node_modules/.bin:$PATH

# install app dependencies
COPY package.json ./
RUN npm install

# add app
COPY . ./

RUN cp tsconfig.standalone.json tsconfig.json

# start app
CMD ["npm", "run", "start-docker"]
