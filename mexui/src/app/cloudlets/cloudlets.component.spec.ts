import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { CloudletsComponent } from './cloudlets.component';

describe('CloudletsComponent', () => {
  let component: CloudletsComponent;
  let fixture: ComponentFixture<CloudletsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ CloudletsComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(CloudletsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
