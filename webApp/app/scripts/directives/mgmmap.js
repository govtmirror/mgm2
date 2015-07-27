'use strict';

function MosesMap(canvas, regions, tiles) {
  var self = this;

  self.centered = false;
  var ctx = canvas.getContext('2d');

  var port = {
    width: canvas.clientWidth,
    height: canvas.clientHeight
  };
  var mouse = {
    down: false,
    x: 0,
    y: 0
  };
  self.offsetX = (port.width / 2) - (256 / 2);
  self.offsetY = (port.height / 2) + (256 / 2);

  var loadedTiles = {}

  var renderingLoop = function () {
    //fill background
    ctx.fillStyle = '#1D475F';
    ctx.fillRect(0, 0, port.width, port.height);

    //draw tile textures
    var offYMod = self.offsetY % 256;
    var offXMod = self.offsetX % 256;
    var width = port.width + 256 * 2;
    var height = port.height + 256 * 2;
    var tileScalar = 256 / self.scalar;
    var x, y, coords, coordstring;
    for (x = offXMod - 256; x < width; x += 256) {
      for (y = offYMod - 256; y < height; y += 256) {
        coords = self.pixelToTile(x, y);
        coordstring = (self.zoomMax - self.zoomLevel + 1) + '-' + coords.x + '-' + coords.y;
        if (coordstring in tiles) {
          if (coordstring in loadedTiles){
            ctx.drawImage(loadedTiles[coordstring], x, y - 256);
          } else {
            var img = new Image();
            img.src = tiles[coordstring];
            img.onload = function(image, url, coordstring){
              loadedTiles[coordstring] = img;
              return function(){
                QueueNewFrame();
              }
            }(img, tiles[coordstring], coordstring);
          }
        }
      }
    }

    //draw names, coordinates, and grid if we are dragging
    if (mouse.down && self.zoomLevel > 5) {
      //draw grid
      ctx.beginPath();
      ctx.lineWidth = 1;
      ctx.strokeStyle = '#777';
      for (x = offXMod - 256; x < width; x += tileScalar) {
        ctx.moveTo(x, 0);
        ctx.lineTo(x, port.height);
      }
      for (y = offYMod - 256; y < height; y += tileScalar) {
        ctx.moveTo(0, y);
        ctx.lineTo(port.width, y);
      }
      ctx.stroke();
      //draw region coordinates and names
      var fontSize = 32 / self.scalar;
      var offset = 256 / self.scalar;
      ctx.font = fontSize + 'px serif';
      ctx.fillStyle = '#FFFFFF';
      for (x = offXMod - 256; x < width; x += tileScalar) {
        for (y = offYMod - 256; y < height; y += tileScalar) {
          coords = self.pixelToTile(x, y);
          coordstring = coords.x + ',' + coords.y;
          ctx.fillText(coordstring, x, y);
          if (coordstring in regions) {
            ctx.fillText(regions[coordstring], x, y - offset + fontSize);
          }
        }
      }
    }
  };

  self.pixelToTile = function (x, y) {
    return {
      'x': Math.floor(-(self.offsetX - x) / (256 / self.scalar)),
      'y': Math.floor((self.offsetY - y) / (256 / self.scalar))
    };
  };

  self.centerTile = function (x, y) {
    self.centered = true;
    self.goToTile(x, y);
    self.offsetX += (port.width / 2) - (256 / 2);
    self.offsetY += (port.height / 2) + (256 / 2);
  };

  self.goToTile = function (x, y) {
    self.offsetX = -x * 256 / self.scalar;
    self.offsetY = y * 256 / self.scalar;
  };

  self.scalar = 1;
  self.zoomLevel = 8;
  self.zoomMin = 1;
  self.zoomMax = 8;
  self.changeZoom = function (delta, x, y) {
    self.zoomLevel += delta;
    if (self.zoomLevel < self.zoomMin) {
      self.zoomLevel = self.zoomMin;
    } else if (self.zoomLevel > self.zoomMax) {
      self.zoomLevel = self.zoomMax;
    }
    //locate tile under mouse
    var parentOffset = canvas.getBoundingClientRect();
    x -= parentOffset.left;
    y -= parentOffset.top;
    var tileCoords = self.pixelToTile(x, y);
    //update scalar
    self.scalar = Math.pow(2, self.zoomMax - self.zoomLevel);
    self.goToTile(tileCoords.x, tileCoords.y);
    //put tile back under mouse
    self.offsetX += x;
    self.offsetY += y;
  };

  canvas.onmousedown = function (e) {
    mouse.down = true;
    mouse.x = e.pageX;
    mouse.y = e.pageY;
    QueueNewFrame();
  };
  canvas.onmouseup = function () {
    mouse.down = false;
    QueueNewFrame();
  };
  canvas.onmouseleave = function () {
    mouse.down = false;
    QueueNewFrame();
  };
  canvas.onmousemove = function (e) {
    if (mouse.down) {
      //drag map
      var dx = e.pageX - mouse.x;
      var dy = e.pageY - mouse.y;
      mouse.x = e.pageX;
      mouse.y = e.pageY;
      self.offsetX += dx;
      self.offsetY += dy;
      QueueNewFrame();
    }
  };
  canvas.onmousewheel = function (e) {
    var delta = Math.max(-1, Math.min(1, (e.wheelDelta || -e.detail)));
    self.changeZoom(delta, e.pageX, e.pageY);
    QueueNewFrame();
  };

  self.resize = function () {
    canvas.width = port.width;
    canvas.height = port.height;
    QueueNewFrame();
  };

  /* tie into web browser refresh mechanisms */
  var intervalID = -1;
  var QueueNewFrame = function () {
      if (window.requestAnimationFrame)
          window.requestAnimationFrame(renderingLoop);
      else if (window.msRequestAnimationFrame)
          window.msRequestAnimationFrame(renderingLoop);
      else if (window.webkitRequestAnimationFrame)
          window.webkitRequestAnimationFrame(renderingLoop);
      else if (window.mozRequestAnimationFrame)
          window.mozRequestAnimationFrame(renderingLoop);
      else if (window.oRequestAnimationFrame)
          window.oRequestAnimationFrame(renderingLoop);
      else {
          QueueNewFrame = function () {
          };
          intervalID = window.setInterval(renderingLoop, 16.7);
      }
  };

  QueueNewFrame();
}



/**
 * @ngdoc directive
 * @name mgmApp.directive:mgmMap
 * @description
 * # mgmMap
 */
angular.module('mgmApp')
  .directive('mgmMap', function () {


    return {
      restrict: 'A',
      link: function postLink(scope, element) {

        var coordTiles = {};
        //map coordinates to regions for name display
        var coordsToRegions = {};
        for (var uuid in scope.regions) {
          var x = scope.regions[uuid].LocX;
          var y = scope.regions[uuid].LocY;
          coordsToRegions[x + ',' + y] = scope.regions[uuid].Name;
          coordTiles['1-' + x + '-' + y] = '/maps/map-1-' + x + '-' + y + '-objects.png';
          for (var z = 2; z <= 8; z++) {
            var w = Math.pow(2, z - 1);
            var x1 = x - (x % w);
            var y1 = y - (y % w);
            coordTiles[z + '-' + x1 + '-' + y1] = '/maps/map-' + z + '-' + x1 + '-' + y1 + '-objects.png';
          }
        }

        var map = new MosesMap(element[0], coordsToRegions, coordTiles);
        map.resize();
        map.centerTile(scope.centerx, scope.centery);

        console.log('New MGM map centered at ' + scope.centerx + ', ' + scope.centery);
      },
      scope: {
        regions: '=',
        centerx: '=',
        centery: '='
      }
    };
  });
