const express = require("express");
const { expressjwt: jwt } = require("express-jwt");
const jwksRsa = require("jwks-rsa");
const cors = require("cors");
const cookieParser = require("cookie-parser");
const crypto = require("crypto");

require("dotenv").config();

const app = express();
const jwksHost = process.env.HANKO_API_URL;

const corsOptions = {
  origin: "http://localhost:8888",
  credentials: true,
  methods: ["GET", "POST", "PATCH", "DELETE", "OPTIONS"],
};

const store = new Map();

app.use(cors(corsOptions));
app.use(cookieParser());
app.use(express.json());
app.use(
  jwt({
    secret: jwksRsa.expressJwtSecret({
      cache: true,
      rateLimit: true,
      jwksRequestsPerMinute: 100,
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
  const userID = req.auth.sub;
  const todos = store.get(userID) || new Map();
  res.status(200).send(Array.from(todos.values()));
});

app.post("/todo", (req, res) => {
  const todoID = crypto.randomBytes(16).toString("hex");
  const { description, checked } = req.body;
  const userID = req.auth.sub;
  const todos = store.get(userID) || new Map();

  todos.set(todoID, { todoID, description, checked });
  store.set(userID, todos);
  res.status(201).end();
});

app.patch("/todo/:id", (req, res) => {
  const userID = req.auth.sub;
  const todoID = req.params.id;
  const todos = store.get(userID);

  if (req.body.hasOwnProperty("checked")) {
    const checked = req.body.checked;

    if (todos.has(todoID)) {
      todos.get(todoID).checked = checked;
    }
  }

  res.status(204).end();
});

app.delete("/todo/:id", (req, res) => {
  const userID = req.auth.sub;
  const todoID = req.params.id;
  const todos = store.get(userID);

  todos.delete(todoID);
  res.status(204).end();
});

app.listen(8002);
