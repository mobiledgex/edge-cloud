# User Database

Users and Organizations are stored in a global sql (postgres) database. This is in contrast to AppInsts and Cloudlets which are stored in Country-specific Etcd databases.

A User represents a person. A User can belong to zero or more Organizations.

An Organization represents an entity such as a Developer Company or Operator Company. However, an Organization could also represent entities within a Company, i.e. Developer Gaming Division vs Developer Machine Learning Division. Organizations are created and managed by Users, so it is up to the user to decide on the scope of the Organization.

An Organization is one of three types, the Admin type (for MobiledgeX support engineers), the Developer type (for Developers), and the Operator type (for Operators). The type of the Organization restricts the type of objects that can managed by that Organization. For example, Developer Organizations can manage Apps, Cluster, but not Cloudlets. Operator Organizations can manage Cloudets, but not Apps and Clusters.

# On-boarding Flow

A person will first access the MobiledgeX website and create a User account. Once their user account is created and confirmed via email (not implemented yet), they can log in and create an Organization. Creating an Organization will automatically add them as the Manager role for that Organization, which has the power to add and remove other users from the Organization.

Once the Organization is created, App, Clusters, and various country-specific data can be created tied to the Organization. Without an Organization, it is not possible to create Apps/Clusters/Cloudlets/etc.

# Frameworks and tools

Postgres is used as the Sql database. Echo is used as the web (REST API) framework. Gorm is used as the ORM database interface.

# Role Based Access

For authorization, we implement role based access using Casbin. Casbin allows for general authorization models. The master controller uses straight forward RBAC except for the fact that developer and operator roles are scoped by Organization, for example a user may be a Developer Admin in one Organization and a Developer Reader in another. This is implemented by prefixing the organization to the user's name in the Casbin profile group (which associates users with roles). Casbin stores the profile configuration into the sql database via an adapater.
