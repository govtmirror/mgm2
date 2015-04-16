'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:GridCtrl
 * @description
 * # GridCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('GridCtrl', function ($scope, mgm, $routeParams) {
    $scope.section = $routeParams["section"];

    if (!$scope.section) {
      $scope.section = "estates";
    }

    $scope.isActive = function (section) {
      return this.section === section;
    }

    $scope.estates = mgm.estates;

    $scope.$on("EstateUpdate", function (event, estate) {
      if (!(estate.ID in $scope.estates)) {
        $scope.estates[estate.ID] = estate;
      } else {
        angular.copy(estate, $scope.estates[estate.ID]);
      }
    });

  });
