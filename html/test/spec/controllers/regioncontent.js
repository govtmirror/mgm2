'use strict';

describe('Controller: RegioncontentCtrl', function () {

  // load the controller's module
  beforeEach(module('mgmApp'));

  var RegioncontentCtrl,
    scope;

  // Initialize the controller and a mock scope
  beforeEach(inject(function ($controller, $rootScope) {
    scope = $rootScope.$new();
    RegioncontentCtrl = $controller('RegioncontentCtrl', {
      $scope: scope
    });
  }));

  it('should attach a list of awesomeThings to the scope', function () {
    expect(scope.awesomeThings.length).toBe(3);
  });
});
