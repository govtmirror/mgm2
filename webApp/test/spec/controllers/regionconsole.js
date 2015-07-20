'use strict';

describe('Controller: RegionconsoleCtrl', function () {

  // load the controller's module
  beforeEach(module('mgmApp'));

  var RegionconsoleCtrl,
    scope;

  // Initialize the controller and a mock scope
  beforeEach(inject(function ($controller, $rootScope) {
    scope = $rootScope.$new();
    RegionconsoleCtrl = $controller('RegionconsoleCtrl', {
      $scope: scope
    });
  }));

  it('should attach a list of awesomeThings to the scope', function () {
    expect(scope.awesomeThings.length).toBe(3);
  });
});
