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
  try {
    const db = loadData();
    res.json(db.manufacturers || []);
  } catch (err) {
    console.error("Error loading manufacturers:", err);
    res.status(500).json({
      error: "Internal Server Error",
      message: "Failed to load manufacturers. Please try again later.",
    });
  }
});

app.get("/api/manufacturers/:id", (req, res) => {
  try {
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
  } catch (err) {
    console.error("Error loading manufacturer:", err);
    res.status(500).json({
      error: "Internal Server Error",
      message: "Failed to load manufacturer. Please try again later.",
    });
  }
});

app.get("/api/categories", (_, res) => {
  try {
    const db = loadData();
    res.json(db.categories || []);
  } catch (err) {
    console.error("Error loading categories:", err);
    res.status(500).json({
      error: "Internal Server Error",
      message: "Failed to load categories. Please try again later.",
    });
  }
});

app.get("/api/categories/:id", (req, res) => {
  try {
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
  } catch (err) {
    console.error("Error loading category:", err);
    res.status(500).json({
      error: "Internal Server Error",
      message: "Failed to load category. Please try again later.",
    });
  }
});

app.get("/api/models", (_, res) => {
  try {
    const db = loadData();
    res.json(db.carModels || []);
  } catch (err) {
    console.error("Error loading models:", err);
    res.status(500).json({
      error: "Internal Server Error",
      message: "Failed to load car models. Please try again later.",
    });
  }
});

app.get("/api/cars", (_, res) => {
  try {
    const db = loadData();
    res.json(db.carModels || []);
  } catch (err) {
    console.error("Error loading cars:", err);
    res.status(500).json({
      error: "Internal Server Error",
      message: "Failed to load cars. Please try again later.",
    });
  }
});

app.get("/api/models/:id", (req, res) => {
  try {
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
  } catch (err) {
    console.error("Error loading model:", err);
    res.status(500).json({
      error: "Internal Server Error",
      message: "Failed to load car model. Please try again later.",
    });
  }
});

app.get("/api/images/:file", (req, res) => {
  try {
    const imagePath = path.join(imagesDir, path.basename(req.params.file));
    if (!fs.existsSync(imagePath)) {
      res.status(404).json({ error: "image not found" });
      return;
    }
    res.sendFile(imagePath);
  } catch (err) {
    console.error("Error serving image:", err);
    res.status(500).json({
      error: "Internal Server Error",
      message: "Failed to serve image. Please try again later.",
    });
  }
});

// Global error handler middleware
app.use((err, req, res, next) => {
  console.error("Server Error:", err);
  res.status(500).json({
    error: "Internal Server Error",
    message: "Something went wrong on the server. Our team has been notified and is working on a fix.",
  });
});

// 404 handler for undefined routes
app.use((req, res) => {
  res.status(404).json({ error: "Route not found" });
});

app.listen(port, () => {
  console.log(`Cars API running on http://localhost:${port}`);
});
