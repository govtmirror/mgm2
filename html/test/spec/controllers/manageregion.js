'use strict';

describe('Controller: ManageregionCtrl', function () {

  // load the controller's module
  beforeEach(module('mgmApp'));

  var ManageregionCtrl,
    scope;

  // Initialize the controller and a mock scope
  beforeEach(inject(function ($controller, $rootScope) {
    scope = $rootScope.$new();
    ManageregionCtrl = $controller('ManageregionCtrl', {
      $scope: scope
    });
  }));

  it('should attach a list of awesomeThings to the scope', function () {
    expect(scope.awesomeThings.length).toBe(3);
  });
});
