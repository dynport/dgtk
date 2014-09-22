# bsondecoder

## Usage

    func run(bsonPath string) error {
      f, e := os.Open(bsonPath)
      if e != nil {
        return e
      }
      defer f.Close()

      dec := bsondecoder.New(f)
      for {
        var m bson.M
        if e = dec.Decode(&m); e != nil {
          return e
        }
        log.Printf("%#v", m)
      }
      return nil
    }
