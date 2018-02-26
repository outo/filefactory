# FileFactory

## What is it?

Provides declarative means to describe (define) filesystem primitives such as Regular files, Directories and Symlinks.
These definitions will be used to create real files and verify that real files exists with specific attributes.  

## Why?

I have been developing some filesystem utilities and found that testing of their intended and side effects
to the filesystem was time consuming, repetitive and the tests were too verbose for my liking.

## Applications

I use it primarily for system testing (involves real filesystem). It is worth bearing in mind that it was
the test case readability, and not the performance that was on my mind when creating it.
In its current shape, I would be careful to use it in production code for the following reasons:
- I haven't evaluated or measured its performance so time critical production applications could suffer
- I'm aware of some limitations for concurrent use. In particular be careful how you use MockForTest function and
avoid using it in production code. This function allows to change the value of package-level variable which can
have undesirable effects if invoked during execution.   

## Mission

My mission was to provide (to myself originally) a declarative testing mechanism to help with:
* defining filesystem primitives, such as regular file, directory and symlink
* defining most useful metainformation about the above file types (aka attributes)
* defining what constitutes a relevant difference between a file definition and a real file (aka verification instruction)
* inspection of the differences between file definition and a real file

## Features

as per contents of examples_test.go and other tests, but few examples are put here for an overview.

### Use file definitions to create and verify real files

*Note: this example shows simple usage case with all the boilerplate included.
As such it does not contribute to better readability of the test.
That will come later, with more complex cases.*

```go
  //first create an instance of FileFactory using constructor function
  ff := filefactory.New()

  //define a regular file, directory and a symlink, use relative path
  fileDefinitions := ff.FilesToCreate(
    def.Reg("relative/path/of/regular-file"),
    def.Dir("relative/path/to/directory"),
    def.Sym("relative/path/to/symlink", "symlink/target/path"),
  )

  //we need some place to create the files (definitions are using relative paths)
  // below we are creating a temporary directory to accommodate them
  tempDirRoot, err := ioutil.TempDir("", "temporary-directory-just-for-this-test")
  if err != nil {
    panic(err)
  }
  //so we don't leave trash behind
  defer os.RemoveAll(tempDirRoot)

  //create files from the above definitions under root
  filefactory.CreateFiles(tempDirRoot, fileDefinitions...)

  // here some processing would occur which shouldn't change these files as a side effect

  //now let's verify files exists and are as defined
  err = filefactory.VerifyFiles(tempDirRoot, fileDefinitions...)
  if err != nil {
    panic(err)
  }

  // they do, because there was no error returned
```

*For brevity, subsequent examples will not include the aforementioned boilerplate. Look at examples_test.go
to see how it can be extracted out of the test cases so it does not get in the way.*

Following snippet is the equivalent of previous one, just without boilerplate and most comments

```go
  fileDefinitions = ff.FilesToCreate(
    def.Reg("relative/path/to/regular-file"),
    def.Dir("relative/path/to/directory"),
    def.Sym("relative/path/to/symlink", "symlink/target/path"),
  )

  filefactory.CreateFiles(tempRootDir, fileDefinitions...)

  //  some processing here

  err := filefactory.VerifyFiles(tempRootDir, fileDefinitions...)
  Expect(err).ShouldNot(HaveOccurred()) //just because I use Gomega but you can use anything else

```
reads better and does not blur the perception of the test criteria.

During testing you can use create, verify or both routines, depending on the goal.
Using both proved extremely helpful when I was testing a recursive copy utility and another, for archiving.

### Use relative paths in file definitions

Relative paths are easier to read than absolute. There is no random looking, long-winded parent directories, just what you want to convey.

My preferred way of testing (as you can see in gomega.BeforeEach functions) is to create a temporary directory
before individual test case and its removal right after the test case. I wouldn't want to prepend all paths with hardcoded paths or variables as it just handicaps the perception.

Using relative paths also means it is easy to reuse the file definitions for both; creating and verifying. You only need to supply root directory path for them routines.

### Create and verify files with non-default attributes

In the above examples at no point was there a mention of any attributes associated with files (files, as a generic filesystem primitive). Each of the primitives defined within this repo can carry a series of attributes or instructions.

An example:
```go
  fileDefinitions = ff.FilesToCreate(
    def.Reg("relative/path/to/regular-file",
      attr.ModePerm(0765),
      attr.OtherGid(),
      attr.ModifiedTime(someTime),
      attr.Size(500),
      attr.Seed(7573453)),
    def.Reg("relative/path/to/another-regular-file",
      attr.ModePerm(0400),
      attr.AccessedTime(someTime),
      attr.Size(200),
      attr.Seed(27)),
    def.Dir("relative/path/to/directory",
      attr.ModePerm(0700),
      attr.AccessedTime(someTime)),
    def.Sym("relative/path/to/symlink", "different/symlink/target/path",
      attr.OtherGid()),
  )

  err := filefactory.CreateFiles(tempRootDir, fileDefinitions...)
  Expect(err).ShouldNot(HaveOccurred())

  // some processing here

  err = filefactory.VerifyFiles(tempRootDir, fileDefinitions...)
  Expect(err).ShouldNot(HaveOccurred())
```

If I wanted to achieve this without this library I obviously could, as there is no magic happening here. However, it would require writing quite a lot of unexciting code.

So, what's available?
- attributes - define the meta information for a file, also can define some aspects of contents. Attributes are used during creation and verification of a file
  - common attributes
    - Owner (uid), can only be used when code runs as a superuser, hence the following helper
      - also CurrentUid() will set the file owner to current user's uid
    - Group (gid),
      - also PrimaryGid() will set the group to current user's primary gid,  
      - also OtherGid() which will select a gid other than primary gid (if available, or primary gid if not)
  - regular files:
    - ModePerm (as in ModePerm bits of os.FileMode) describes file's permissions. Value of mode type equivalent on the other hand, is controlled within function creating `DefinitionConstructor`. The values provided to ModePerm are most recognizable when typically specified as octal (i.e. in Go preceded by zero).
    - Size - will create an actual file of that length, it will be populated with pseudo-random (Seed) bytes' sequence
    - Seed - this is a concept introduced by me. Setting the seed and size same on two files will produce files' contents with equal byte sequence. That allowed me testing for any form of data corruption during manipulating file contents.
    - Modified time
    - Accessed time
  - directories:
    - ModePerm (as in ModePerm bits of os.FileMode) describes file's permissions. Value of mode type equivalent on the other hand, is controlled within function creating `DefinitionConstructor`. The values provided to ModePerm are most recognizable when typically specified as octal (i.e. in Go preceded by zero).
    - Modified time
    - Accessed time
  - symlinks only:
    - Link target - this is not settable via the attribute mechanism but on the constructor
- verification instructions - a set of boolean-like constructs which define the behaviour of verification
  - AllByDefault - this is a default for the whole of verification for that file or file factory (depending where it is set)
  - ModePerm - turn on/off verification of file's permissions (handy for symlinks where mode does not make much sense)
  - ModifiedTime - could be useful if you are verifying existence of a directory into which file has been added
  - AccessedTime - this was useful for tar headers which reset the accessed time
  - Uid - this is only useful when code runs with root
  - Gid
  - Size - it does not read the file, just retrieves Size from os.FileInfo
  - SymlinkTarget - will check value of immediate target
  - Contents - in the Regular, it will read the contents and compare to in-memory contents created by using Seed and Size attributes

*Note: Is is possible to extend the functionality of this library, including file primitives, attributes and verification instructions. Have a look at `def` pkg. It contains a file per each primitive. This file is able to handle the specifics of creating a definition, creating a real-life equivalent and verifying it's existence and attributes.*

### Inspect the errors

Let's check a file does not exist:

```go
  filesToExpect := ff.FilesToExpect(
  	def.Reg("non/existent/file"),
  )

  err := fileFactory.VerifyFiles(tempRootDir, filesToExpect...)

  //omitted nil and type check (can panic)

  verErr := err.(*ff.VerificationErrors)
  Expect(verErr.HasDifference(diff.NotPresentOrNotAccessible, abs("non/existent/file")))
```

where `abs()` is just a locally defined convenience:
```go
  abs := func(relPath string) (absPath string) {
    return filepath.Join(tempRootDir, relPath)
  }
```

*Note: `NotPresentOrNotAccessible` will always be returned in case of non-existent or inaccessible paths. `DiffModeType` will always be returned in case file type (but not ModePerm) is not aligned with definition. `DiffModePerm` is switchable though. Both differences will interrupt the verification process for that file.*

Demo variety of assertions (you don't need to do all of this)
```go
  filesToCreate := fileFactory.FilesToCreate(
    def.Reg("relative/path/to/regular-file"),
    def.Dir("relative/path/to/directory"),
    def.Sym("relative/path/to/symlink", "symlink/target/path"),
  )

  err := filefactory.CreateFiles(tempRootDir, filesToCreate...)
  Expect(err).ShouldNot(HaveOccurred())

  //the following declarations are for the same file paths as above but with different attributes
  verifyDeclarations := fileFactory.FilesToExpect(
    def.Reg("relative/path/to/regular-file", attr.OtherGid(), attr.ModePerm(0765), attr.ModifiedTime(time.Now())),
    def.Dir("relative/path/to/directory", attr.ModePerm(0700), attr.AccessedTime(time.Now())),
    def.Sym("relative/path/to/symlink", "different/symlink/target/path"),
  )

  err = filefactory.VerifyFiles(tempRootDir, verifyDeclarations...)
  Expect(err).Should(HaveOccurred())
  Expect(err).To(BeAssignableToTypeOf(&verify.Errors{}))
  verErr := err.(*verify.Errors)

  //this check gives me indication if particular type of error happened at least once for any of the declarations
  Expect(verErr.CombinedFileDifference & diff.ModePerm).To(Equal(diff.ModePerm))

  //but because each of the registered difference types takes a single bit we can say
  Expect(verErr.CombinedFileDifference & diff.ModePerm).ToNot(BeZero())

  //this tells me that each of the differences was detected
  Expect(verErr.CombinedFileDifference).To(Equal(diff.ModePerm | diff.ModTime | diff.AccTime | diff.Group | diff.LinkTarget))

  //and this is detailed list of the error type, path and original error itself (similar to os.PathError)
  Expect(verErr.Errors).To(HaveLen(6))
  Expect(verErr.HasDifference(diff.Group, abs("relative/path/to/regular-file"))).To(BeTrue())
  Expect(verErr.HasDifference(diff.ModePerm, abs("relative/path/to/regular-file"))).To(BeTrue())
  Expect(verErr.HasDifference(diff.ModTime, abs("relative/path/to/regular-file"))).To(BeTrue())
  Expect(verErr.HasDifference(diff.ModePerm, abs("relative/path/to/directory"))).To(BeTrue())
  Expect(verErr.HasDifference(diff.AccTime, abs("relative/path/to/directory"))).To(BeTrue())
  Expect(verErr.HasDifference(diff.LinkTarget, abs("relative/path/to/symlink"))).To(BeTrue())

  //also we can verify what differences were registered against a particular path
  Expect(verErr.DifferenceFor(abs("relative/path/to/directory"))).To(Equal(diff.ModePerm | diff.AccTime))
```

In case an attribute isn't explicitly passed in, the default value will take precedence.

### Attribute precedence

There are few factors determining what attribute will be set on files. Attributes with lower precedence will be overwritten. In the order of lowest to highest precedence:
- hardcoded defaults:
  - the weakest attribute is the `hardcodedFileFactoryDefaults` and it is located in FileFactory. Why?:
    - it fixes the uid and gid of a file to the safest option, which I assume is current user's uid and primary gid, otherwise unnecessary invocations of chown would be triggerred which in turn will fail if code does not run from root account
    - it fixes the AccessedTime and ModifiedTime timestamps on a file so that all file definitions created with same instance of a factory will carry the same default timestamps (but different for AccessedTime and ModifiedTime), if your tests requires verification of any of them just provide the timestamp at the factory or definition level
  - next up is `fileSpecificDefaults` which is different for each of the filesystem primitives and can be found in functions creating `DefinitionConstructor`s
    - def.Reg() - set default file mode to something usable and set the size to non-zero so we can more easily spot data corruption
    - def.Dir() - set default file mode to something usable
    - def.Sym() - no attributes but there are verification instructions preventing verification of mode and timestamps which I don't believe play role in symlinks
- specified by the user:
  - `extraFileFactoryDefaults` which are provided at the time of FileFactory creation will affect each of the file definitions created with this factory instance
  - finally, the highest precedence override for `extraFileSpecificAttributes` which is specified next to file definition
*Note: if you are creating new filesystem primitive type you can decide on different precedence as the order of them would be defined in equivalent of Reg(), Dir or  Sym() that you would write anyway.*


```go
  //create file factory which only verifies mode by default
  fileFactory = filefactory.New(verify.AllByDefault(false), verify.ModePerm(true))

  fileDefinitions := fileFactory.FilesToCreate(
  	def.Reg("relative/path/to/regular-file"),
  	def.Dir("relative/path/to/directory"),
  	def.Sym("relative/path/to/symlink", "symlink/target/path"),
  )

  //these files will have incorrect attributes
  sameFilesButDifferentAttributes := fileFactory.FilesToExpect(
  	def.Reg("relative/path/to/regular-file", attr.OtherGid(), attr.ModePerm(0765), attr.ModifiedTime(time.Now())),
  	def.Dir("relative/path/to/directory", attr.ModePerm(0700), attr.AccessedTime(time.Now())),
  	//for symlinks we have to override the the mode verification as it does not make sense
  	def.Sym("relative/path/to/symlink", "different/symlink/target/path", verify.ModePerm(false)),
  )

  err := filefactory.CreateFiles(tempRootDir, fileDefinitions...)
  Expect(err).ShouldNot(HaveOccurred())

  err = filefactory.VerifyFiles(tempRootDir, sameFilesButDifferentAttributes...)
  Expect(err).Should(HaveOccurred())
  Expect(err).To(BeAssignableToTypeOf(&verify.Errors{}))
  verErr := err.(*verify.Errors)

  // the only observable errors are the ones to do with regular file and directory's mode
  Expect(verErr.Errors).To(HaveLen(2))
  Expect(verErr.HasDifference(diff.ModePerm, abs("relative/path/to/regular-file")))
  Expect(verErr.HasDifference(diff.ModePerm, abs("relative/path/to/directory")))
```

In this example factory was created so that it does not validate much more than file permissions (file presence and file type is always checked).
That will apply to all definitions created with this factory. However, for symlink, normally I don't want to check the permissions.
I can override the factory-level value with definition-level value just for this one definition.

## Running tests

I am using Ginkgo and Gomega for testing.
To install them execute:
```
$ go get github.com/onsi/ginkgo/ginkgo
$ go get github.com/onsi/gomega/...
```

Change the working directory to the root of this repo and use

```
$ ginkgo -r --randomizeAllSpecs --randomizeSuites --failOnPending --cover --trace --race --progress -p
```

## Compatibility

It is currently only compatible with Unix os family. I haven't got a plan to make it work with Windows or any other platforms as I have no use for that and little spare time.

In terms of testing framework this library is not dependent on any specific one. I use Ginkgo/Gomega because I quite like it.

## Stability of this repo

I cannot guarantee stability as it is fairly fresh package and I expect further commits, some of which may violate the interface. Please fork it.

## More examples

Please review the tests, they document all features.