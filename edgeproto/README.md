# Protocol Buffers Definition Files

Both gogo proto and MEX custom extensions are used.

Gogo proto buffer extension to embed struct as a value rather than a pointer is necessary for object key structs that are composed of other key structs. These key structs are used as the hash key for maps internally, so cannot have pointer references. I suppose the maps could have converted the key structs to strings first, but that would add unnecessary overhead.

Our custom extensions are for handling database objects. A database object has an embedded key that uniquely defines it. It also has a fields field to specify which fields are to update on the update call. An object should have an {{ObjName}}Api service which defines a Create, Delete, Update, and Show (stream output) set of commands. With these requirements met, the protogen.generate_cud annotation can be added and a bunch of support (and test) code will be automatically generated. Failure to set up the struct with the correct fields will likely result in compile-time failures if protogen.generate_cud is used.
