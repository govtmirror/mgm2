'use strict';

describe('Service: MgmPublic', function () {

  // load the service's module
  beforeEach(module('mgmApp'));

  // instantiate service
  var MgmPublic;
  beforeEach(inject(function (_MgmPublic_) {
    MgmPublic = _MgmPublic_;
  }));

  it('should do something', function () {
    expect(!!MgmPublic).toBe(true);
  });

});
