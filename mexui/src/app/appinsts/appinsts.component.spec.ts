import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { AppinstsComponent } from './appinsts.component';

describe('AppinstsComponent', () => {
  let component: AppinstsComponent;
  let fixture: ComponentFixture<AppinstsComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ AppinstsComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(AppinstsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
