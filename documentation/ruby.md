To use Chew with Ruby you need to first build the package to create the shared object file and header file. You can do this by running the following command:

```bash
go build -o chew.so -buildmode=c-shared ./cmd/chew/wrapper.go
```

This will create a `chew.so` and `chew.h` file in the current directory. You can then use these files in your Ruby project to use Chew. Here is an example of how to use Chew in your Ruby project:

```ruby
require 'fiddle'
require 'fiddle/import'

module ChewLib
  extend Fiddle::Importer
  dlload './chew.so'

  extern 'char* Process(char*)'
end

urls = ['https://example.com', 'https://example.com']
for url in urls
  result_ptr = ChewLib.Process(url)
  result = result_ptr.to_s
  Fiddle::Function.new(Fiddle::Handle['free'], [Fiddle::TYPE_VOIDP], Fiddle::TYPE_VOID).call(result_ptr)

  puts result
end
```

Using chew like this will come with obvious limitations, however, this is a simple example of how to use Chew in your Ruby project
