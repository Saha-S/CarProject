const express = require("express");
const fs = require("fs");
const path = require("path");

const app = express();
const port = process.env.API_PORT || 3000;
const dataPath = path.join(__dirname, "data.json");
const imagesDir = path.join(__dirname, "images");

function loadData() {
  const raw = fs.readFileSync(dataPath, "utf8");
  return JSON.parse(raw);
}

function toInt(value) {
  const n = Number.parseInt(value, 10);
  return Number.isNaN(n) ? null : n;
}

app.get("/api/health", (_, res) => {
  res.json({ ok: true });
});

app.get("/api/manufacturers", (_, res) => {
  const db = loadData();
  res.json(db.manufacturers || []);
});

app.get("/api/manufacturers/:id", (req, res) => {
  const db = loadData();
  const id = toInt(req.params.id);
  if (id === null) {
    res.status(400).json({ error: "invalid id" });
    return;
  }
  const manufacturer = (db.manufacturers || []).find((m) => m.id === id);
  if (!manufacturer) {
    res.status(404).json({ error: "not found" });
    return;
  }
  res.json(manufacturer);
});

app.get("/api/categories", (_, res) => {
  const db = loadData();
  res.json(db.categories || []);
});

app.get("/api/categories/:id", (req, res) => {
  const db = loadData();
  const id = toInt(req.params.id);
  if (id === null) {
    res.status(400).json({ error: "invalid id" });
    return;
  }
  const category = (db.categories || []).find((c) => c.id === id);
  if (!category) {
    res.status(404).json({ error: "not found" });
    return;
  }
  res.json(category);
});

app.get("/api/models", (_, res) => {
  const db = loadData();
  res.json(db.carModels || []);
});

app.get("/api/cars", (_, res) => {
  const db = loadData();
  res.json(db.carModels || []);
});

app.get("/api/models/:id", (req, res) => {
  const db = loadData();
  const id = toInt(req.params.id);
  if (id === null) {
    res.status(400).json({ error: "invalid id" });
    return;
  }
  const model = (db.carModels || []).find((c) => c.id === id);
  if (!model) {
    res.status(404).json({ error: "not found" });
    return;
  }
  res.json(model);
});

app.get("/api/images/:file", (req, res) => {
  const imagePath = path.join(imagesDir, path.basename(req.params.file));
  if (!fs.existsSync(imagePath)) {
    res.status(404).json({ error: "image not found" });
    return;
  }
  res.sendFile(imagePath);
});

app.listen(port, () => {
  console.log(`Cars API running on http://localhost:${port}`);
});
