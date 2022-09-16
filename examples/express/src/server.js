const express = require("express");
const { expressjwt: jwt } = require("express-jwt");
const jwksRsa = require("jwks-rsa");
const cors = require("cors");
const cookieParser = require("cookie-parser");

require("dotenv").config();

const app = express();
const jwksHost = process.env.HANKO_API_URL;

const corsOptions = {
  origin: "http://localhost:3000",
  credentials: true,
  methods: ["GET", "POST", "PATCH", "DELETE", "OPTIONS"],
};

const store = {};

app.use(cors(corsOptions));
app.use(cookieParser());
app.use(express.json());
app.use(
  jwt({
    secret: jwksRsa.expressJwtSecret({
      cache: true,
      rateLimit: true,
      jwksRequestsPerMinute: 2,
      jwksUri: `${jwksHost}/.well-known/jwks.json`,
    }),
    algorithms: ["RS256"],
    getToken: function fromCookieOrHeader(req) {
      if (
        req.headers.authorization &&
        req.headers.authorization.split(" ")[0] === "Bearer"
      ) {
        return req.headers.authorization.split(" ")[1];
      } else if (req.cookies && req.cookies.hanko) {
        return req.cookies.hanko;
      }
      return null;
    },
  })
);

app.get("/todo", (req, res) => {
  res.status(200).send(store[req.auth.sub] || []);
});

app.post("/todo", (req, res) => {
  const { description, checked } = req.body;
  (store[req.auth.sub] ||= []).push({ description, checked });
  res.status(201).end();
});

app.patch("/todo/:id", (req, res) => {
  if (req.body.hasOwnProperty("checked")) {
    store[req.auth.sub][req.params.id].checked = req.body.checked;
  }
  res.status(204).end();
});

app.delete("/todo/:id", (req, res) => {
  store[req.auth.sub].splice(req.params.id, 1);
  res.status(204).end();
});

app.get("/logout", (req, res) => {
  res.clearCookie("hanko");
  res.status(204).end();
});

app.listen(8002);
