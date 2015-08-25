'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:RegionconsoleCtrl
 * @description
 * # RegionconsoleCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('RegionconsoleCtrl', function ($scope, $modalInstance, $timeout, mgm, region) {

    $scope.region = region;
    $scope.lines = []

    $scope.close = function () {
      $modalInstance.close();
    };

    $scope.$on('ConsoleInput', function (event, line) {
      mgm.request('ConsoleCommand', {
        Message: line
      }, function (success, msg) {
        if (success) {
          //alertify.success(msg);
        } else {
          alertify.error(msg);
        }
      });
    });

    $scope.$on('ConsoleUpdate', function (event, status) {
      if (status.UUID == region.UUID) {
        var lines = []
        for (var i = 0; i < status.Lines.length; i++) {
          var line = status.Lines[i];
          console.log(line);
          var parts = line.split(':')
          var number = parts.shift();
          var level = parts.shift();
          line = parts.join(':');
          //if (line.startsWith('+++Region ') || line === '' ) {
          //  //skip the faux prompts
          //  continue;
          //}
          parts = line.split('\n');
          for (var j = 0; j < parts.length; j++) {
            lines.push({
              Number: number,
              Level: level,
              Message: parts[j]
            });
          }
        }
        if (lines.length > 0) {
          $timeout(function () {
            $scope.lines.push.apply($scope.lines, lines);
          });
        }
      }
    });

  });
