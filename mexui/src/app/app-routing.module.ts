import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { DevelopersComponent } from './developers/developers.component';
import { AppsComponent } from './apps/apps.component';
import { AppinstsComponent } from './appinsts/appinsts.component';
import { CloudletsComponent } from './cloudlets/cloudlets.component';
import { ClustersComponent } from './clusters/clusters.component';
import { FlavorsComponent } from './flavors/flavors.component';
import { OperatorsComponent } from './operators/operators.component';
const routes: Routes = [
  {
    path: 'Developers',
    component: DevelopersComponent
  },
  {
    path: 'Apps',
    component: AppsComponent
  },
  {
    path: 'Appinsts',
    component: AppinstsComponent
  },
  {
    path: 'Flavors',
    component: FlavorsComponent
  },
  {
    path: 'Cloudlets',
    component: CloudletsComponent
  },
  {
    path: 'Clusters',
    component: ClustersComponent
  },
  {
    path: 'Operators',
    component: OperatorsComponent
  },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
