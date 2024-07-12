# please see the documentation on how to build chew for use with ruby

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
