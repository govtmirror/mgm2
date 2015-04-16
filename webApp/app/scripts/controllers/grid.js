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

    $scope.getUserNameFromID = function (uuid) {
      if (uuid in mgm.users) {
        return mgm.users[uuid].Name
      }
      return ""
    }
    $scope.getUserNamesFromIDs = function (uuids) {
      var users = [];
      for (var i = 0; i < uuids.length; i++) {
        var uuid = uuids[i];
        if (uuid in mgm.users) {
          users.push(mgm.users[uuid].Name);
        }
      }
      return users;
    }
  });
