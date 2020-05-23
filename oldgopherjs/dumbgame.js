document.addEventListener("DOMContentLoaded", init);

var CSS_PIXEL_WIDTH = window.devicePixelRatio;
var CSS_PIXEL_OFFSET = CSS_PIXEL_WIDTH / 2;
var PHYSICAL_CSS_PIXEL_OFFSET = 0.5;

var KEYCODE_LEFT = 37;
var KEYCODE_UP = 38;
var KEYCODE_RIGHT = 39;
var KEYCODE_DOWN = 40;

var game = null;

function init(event) {
  var canvas = document.getElementById("canvas");
  var ctx = canvas.getContext("2d");

  console.log("window.devicePixelRatio:", window.devicePixelRatio);

  // fun hack to work around high DPI displays
  var originalWidth = canvas.width;
  var originalHeight = canvas.height;
  canvas.width = originalWidth * window.devicePixelRatio;
  canvas.height = originalHeight * window.devicePixelRatio;
  canvas.style.width = originalWidth + "px";
  canvas.style.height = originalHeight + "px";

  // draw(ctx, 300, 300);

  game = dumbgame.NewGame(ctx);
  // TODO: Move this initialization into Go?
  document.addEventListener("keydown", game.KeyDown);
  document.addEventListener("keyup", game.KeyUp);
}

function draw(ctx, x, y) { 
  ctx.fillStyle = "green";
  ctx.fillRect(10, 10, 100, 100);

  var SIZE = 10 * CSS_PIXEL_WIDTH;
  var OFFSET = SIZE / 2 + CSS_PIXEL_OFFSET;
  ctx.lineWidth = CSS_PIXEL_WIDTH;
  ctx.strokeStyle = "black";
  ctx.strokeRect(x - OFFSET, y - OFFSET, SIZE, SIZE);
  ctx.beginPath();
  console.log("line", x - CSS_PIXEL_OFFSET, y - CSS_PIXEL_OFFSET, x - CSS_PIXEL_OFFSET, y-SIZE - CSS_PIXEL_OFFSET);
  ctx.moveTo(x - CSS_PIXEL_OFFSET, y - CSS_PIXEL_OFFSET);
  ctx.lineTo(x - CSS_PIXEL_OFFSET, y-SIZE - CSS_PIXEL_OFFSET);
  ctx.stroke();
}