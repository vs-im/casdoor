const fs = require("fs");
const path = require("path");

const sourceDir = path.join(__dirname, "build-temp");
const targetDir = path.join(__dirname, "build");

if (!fs.existsSync(sourceDir)) {
  // eslint-disable-next-line no-console
  console.error(`Source directory "${sourceDir}" does not exist.`);
  process.exit(1);
}

if (fs.existsSync(targetDir)) {
  fs.rmSync(targetDir, { recursive: true, force: true });
  // eslint-disable-next-line no-console
  console.log(`Target directory "${targetDir}" has been deleted successfully.`);
}

try {
  fs.renameSync(sourceDir, targetDir);
} catch (err) {
  if (err.code === "EXDEV") {
    fs.cpSync(sourceDir, targetDir, { recursive: true });
    fs.rmSync(sourceDir, { recursive: true, force: true });
  } else {
    throw err;
  }
}
// eslint-disable-next-line no-console
console.log(`Renamed "${sourceDir}" to "${targetDir}" successfully.`);
