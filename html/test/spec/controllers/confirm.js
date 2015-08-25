'use strict';

describe('Controller: ConfirmCtrl', function () {

  // load the controller's module
  beforeEach(module('mgmApp'));

  var ConfirmCtrl,
    scope;

  // Initialize the controller and a mock scope
  beforeEach(inject(function ($controller, $rootScope) {
    scope = $rootScope.$new();
    ConfirmCtrl = $controller('ConfirmCtrl', {
      $scope: scope
    });
  }));

  it('should attach a list of awesomeThings to the scope', function () {
    expect(scope.awesomeThings.length).toBe(3);
  });
});
