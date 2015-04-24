'use strict';

/**
 * @ngdoc directive
 * @name mgmApp.directive:cpuChart
 * @description
 * # mgmHostRow
 */
angular.module('mgmApp')
  .directive('mgmChart', function ($timeout) {

    var template = '';

    function type(d) {
      d.value = +d.value; // coerce to number
      return d;
    }

    var drawChart = function (scope, element, attrs, dataset) {
      var id = attrs.id
      var margin = {
        top: 0,
        right: 0,
        bottom: 0,
        left: 0
      };
      var width = 80 - margin.left - margin.right;
      var height = 40 - margin.top - margin.bottom;
      var w = width;
      var h = height;


      var xScale = d3.scale.ordinal()
        .domain(dataset.map(function (d) {
          return d.key;
        }))
        .rangeRoundBands([margin.left, width], 0.40);

      var yScale = d3.scale.linear()
        .domain([0, d3.max(dataset, function (d) {
          return d.val;
        })])
        .range([h, 0]);

      var svg = d3.select('#' + id).append("svg")
        .attr("width", w + margin.right + margin.left)
        .attr("height", h + margin.top + margin.bottom)

      svg.selectAll("rect")
        .data(dataset)
        .enter().append("rect")
        .attr("x", function (d, i) {
          return xScale(d.key);
        })
        .attr("y", function (d) {
          return yScale(d.val);
        })
        .attr("width", xScale.rangeBand())
        .attr("height", function (d) {
          return h - yScale(d.val);
        })
        .attr("fill", function (d) {
          return "rgb(259, 148, " + (d.val * 14) + ")";
        })
    }

    var linkFunction = function (scope, element, attrs) {
      var data = scope.data;
      if (data === undefined) {
        return;
      }

      console.log(data);

      var graphData = [];
      for (var i = 0; i < data.length; i++) {
        if (data[i] < 2) {
          data[i] = 2;
        }
        graphData.push({
          'key': i,
          'val': data[i]
        });
      }

      $timeout(function () {
        drawChart(scope, element, attrs, graphData);
      }, 1);

    };

    return {
      template: template,
      link: linkFunction,
      scope: {
        data: '=',
      }
    };
  });
