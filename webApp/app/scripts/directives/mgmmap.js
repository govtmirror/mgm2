'use strict';

/**
 * @ngdoc directive
 * @name mgmApp.directive:mgmMap
 * @description
 * # mgmMap
 */
angular.module('mgmApp')
  .directive('mgmMap', function () {


    return {
      template: '<canvas id="mosesMap" style="width: 100%; height: 100%"></div>',
      restrict: 'A',
      link: function postLink(scope, element, attrs) {

        var coordTiles = {};
        //map coordinates to regions for name display
        var coordsToRegions = {};
        for (var uuid in scope.regions) {
          var x = scope.regions[uuid].LocX;
          var y = scope.regions[uuid].LocY;
          coordsToRegions[x + "," + y] = scope.regions[uuid].Name;
          coordTiles["1-" + x + "-" + y] = "/maps/map-1-" + x + "-" + y + "-objects.png";
        }

        //map

        var canvas = document.getElementById('mosesMap');
        var map = new MosesMap(canvas, coordsToRegions, coordTiles);
        map.resize();
        map.centerTile(1000, 1000);

      },
      scope: {
        regions: '='
      }
    };
  });


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
  }
  self.offsetX = (port.width / 2) - (256 / 2);
  self.offsetY = (port.height / 2) + (256 / 2);

  self.redraw = function () {
    ctx.fillStyle = "#1D475F";
    ctx.fillRect(0, 0, port.width, port.height);
    self.drawGrid();
  }

  self.pixelToTile = function (x, y) {
    return {
      'x': Math.floor(-(self.offsetX - x) / (256 / self.scalar)),
      'y': Math.floor((self.offsetY - y) / (256 / self.scalar))
    };
  }

  self.centerTile = function (x, y) {
    self.centered = true;
    self.goToTile(x, y);
    self.offsetX += (port.width / 2) - (256 / 2);
    self.offsetY += (port.height / 2) + (256 / 2);
  }

  self.goToTile = function (x, y) {
    self.offsetX = -x * 256 / self.scalar;
    self.offsetY = y * 256 / self.scalar;
  }

  self.drawGrid = function () {
    var offYMod = self.offsetY % 256;
    var offXMod = self.offsetX % 256;
    var width = port.width + 256 * 2;
    var height = port.height + 256 * 2;
    var tileScalar = 256 / self.scalar;
    for (var x = offXMod - 256; x < width; x += 256) {
      for (var y = offYMod - 256; y < height; y += 256) {
        var coords = self.pixelToTile(x, y);
        var coordstring = self.scalar + "-" + coords.x + "-" + coords.y;
        var tileName = "map-" + (self.zoomMax - self.zoomLevel + 1) + "-" + coords.x + "-" + coords.y + "-objects.png";
        console.log(tileName);
        if (coordstring in tiles) {
          var img = new window.Image();
          img.setAttribute("src", tiles[coordstring]);
          ctx.drawImage(img, x, y);
        }
      }
    }
    if (mouse.down && self.zoomLevel > 4) {
      //draw grid
      ctx.beginPath();
      ctx.lineWidth = 1;
      ctx.strokeStyle = "#777";
      for (var x = offXMod - 256; x < width; x += tileScalar) {
        ctx.moveTo(x, 0);
        ctx.lineTo(x, port.height);
      }
      for (var y = offYMod - 256; y < height; y += tileScalar) {
        ctx.moveTo(0, y);
        ctx.lineTo(port.width, y);
      }
      ctx.stroke();
      //draw region coordinates and names
      var fontSize = 32 / self.scalar;
      var offset = 256 / self.scalar;
      ctx.font = fontSize + "px serif";
      ctx.fillStyle = "#FFFFFF";
      for (var x = offXMod - 256; x < width; x += tileScalar) {
        for (var y = offYMod - 256; y < height; y += tileScalar) {
          var coords = self.pixelToTile(x, y);
          var coordstring = coords.x + "," + coords.y;
          ctx.fillText(coordstring, x, y + offset);
          if (coordstring in regions) {
            ctx.fillText(regions[coordstring], x, y + fontSize);
          }
        }
      }
    }
  }

  self.scalar = 1;
  self.zoomLevel = 8;
  self.zoomMin = 1;
  self.zoomMax = 8;
  self.changeZoom = function (delta, x, y) {
    self.zoomLevel += delta;
    if (self.zoomLevel < self.zoomMin) {
      self.zoomLevel = self.zoomMin;
      return;
    } else if (self.zoomLevel > self.zoomMax) {
      self.zoomLevel = self.zoomMax;
      return;
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
  }

  canvas.onmousedown = function (e) {
    mouse.down = true
    mouse.x = e.pageX;
    mouse.y = e.pageY;
    self.redraw();
  };
  canvas.onmouseup = function () {
    mouse.down = false;
    self.redraw();
  };
  canvas.onmouseleave = function () {
    mouse.down = false;
    self.redraw();
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
      self.redraw();
    }
  };
  canvas.onmousewheel = function (e) {
    var delta = Math.max(-1, Math.min(1, (e.wheelDelta || -e.detail)));
    self.changeZoom(delta, e.pageX, e.pageY);
    self.adjustOffset
    self.redraw();
  };

  self.resize = function () {
    canvas.width = port.width;
    canvas.height = port.height;
    self.redraw();
  }
};
