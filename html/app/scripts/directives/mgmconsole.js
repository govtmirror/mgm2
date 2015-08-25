'use strict';

/**
 * @ngdoc directive
 * @name mgmApp.directive:mgmConsole
 * @description
 * # mgmConsole
 */
angular.module('mgmApp')
  .directive('mgmConsole', function ($rootScope, $timeout) {

    var linkFunction = function (scope, element, attrs) {
      scope.input = '';

      //store commands for console-style up-down navigation
      var currentCommand = 1;
      var commandsList = [''];

      scope.keyUp = function ($event) {

        switch ($event.keyCode) {
        case 13: //return
          $timeout(function () {
            commandsList[commandsList.length-1] = scope.input;
            $rootScope.$broadcast('ConsoleInput', scope.input);
            scope.input = '';
            commandsList.push(scope.input);
            currentCommand = 1;
          })
          return;
          break;
        case 38: //up arrow
          currentCommand +=1;
          if(currentCommand >= commandsList.length){
            currentCommand = commandsList.length;
          }
          $timeout(function () {
            scope.input = commandsList[commandsList.length-currentCommand];
          })
          return;
          break;
        case 40: //down arrow
          currentCommand -= 1;
          if(currentCommand <= 1){
            currentCommand = 1;
          }
          $timeout(function () {
            scope.input = commandsList[commandsList.length-currentCommand];
          })
          return;
          break;
        default:
          commandsList[commandsList.length-1] = scope.input;
        }

        //console.log($event);
        //$event.preventDefault();
      };
    };

    return {
      template: '<div ng-keyup="keyUp($event)">' +
        '  <div class="console-window" scroll-glue-bottom>' +
        '    <ul>' +
        '      <li ng-repeat="line in lines track by $index">{{line.Message}}</li>' +
        '    </ul>' +
        '  </div>' +
        '  <div class="console-prompt"> $> </div>' +
        '  <input class="console-input" ng-model="input" autofocus></span>' +
        '  </div>' +
        '</div',
      //restrict: 'E',
      link: linkFunction,
      scope: {
        lines: '=',
      }
    };
  });
