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
      message: "We're having trouble loading the manufacturer data. Please try again shortly.",
    });
  }
});

app.get("/api/manufacturers/:id", (req, res) => {
  try {
    const db = loadData();
    const id = toInt(req.params.id);
    if (id === null) {
      res.status(400).json({ message: "Invalid manufacturer ID." });
      return;
    }
    const manufacturer = (db.manufacturers || []).find((m) => m.id === id);
    if (!manufacturer) {
      res.status(404).json({ message: "Manufacturer not found." });
      return;
    }
    res.json(manufacturer);
  } catch (err) {
    console.error("Error loading manufacturer:", err);
    res.status(500).json({
      message: "We're having trouble loading manufacturer details. Please try again shortly.",
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
      message: "We're having trouble loading the category data. Please try again shortly.",
    });
  }
});

app.get("/api/categories/:id", (req, res) => {
  try {
    const db = loadData();
    const id = toInt(req.params.id);
    if (id === null) {
      res.status(400).json({ message: "Invalid category ID." });
      return;
    }
    const category = (db.categories || []).find((c) => c.id === id);
    if (!category) {
      res.status(404).json({ message: "Category not found." });
      return;
    }
    res.json(category);
  } catch (err) {
    console.error("Error loading category:", err);
    res.status(500).json({
      message: "We're having trouble loading category details. Please try again shortly.",
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
      message: "We're having trouble loading the car models. Please try again shortly.",
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
      message: "We're having trouble loading the cars. Please try again shortly.",
    });
  }
});

app.get("/api/models/:id", (req, res) => {
  try {
    const db = loadData();
    const id = toInt(req.params.id);
    if (id === null) {
      res.status(400).json({ message: "Invalid car ID." });
      return;
    }
    const model = (db.carModels || []).find((c) => c.id === id);
    if (!model) {
      res.status(404).json({ message: "Car not found." });
      return;
    }
    res.json(model);
  } catch (err) {
    console.error("Error loading model:", err);
    res.status(500).json({
      message: "We're having trouble loading car details. Please try again shortly.",
    });
  }
});

app.get("/api/images/:file", (req, res) => {
  try {
    const imagePath = path.join(imagesDir, path.basename(req.params.file));
    if (!fs.existsSync(imagePath)) {
      res.status(404).json({ message: "Image not found." });
      return;
    }
    res.sendFile(imagePath);
  } catch (err) {
    console.error("Error serving image:", err);
    res.status(500).json({
      message: "We're having trouble loading the image. Please try again shortly.",
    });
  }
});

// Global error handler middleware
app.use((err, req, res, next) => {
  console.error("Server Error:", err);
  res.status(500).json({
    message: "Something went wrong. Please refresh the page and try again.",
  });
});

// 404 handler for undefined routes
app.use((req, res) => {
  res.status(404).json({ message: "The requested resource was not found." });
});

app.listen(port, () => {
  console.log(`Cars API running on http://localhost:${port}`);
});
