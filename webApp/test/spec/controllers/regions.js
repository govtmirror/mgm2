'use strict';

describe('Controller: RegionsCtrl', function () {

  // load the controller's module
  beforeEach(module('mgmApp'));

  var RegionsCtrl,
    scope;

  // Initialize the controller and a mock scope
  beforeEach(inject(function ($controller, $rootScope) {
    scope = $rootScope.$new();
    RegionsCtrl = $controller('RegionsCtrl', {
      $scope: scope
    });
  }));

});
