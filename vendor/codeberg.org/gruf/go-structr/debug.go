package structr

// String returns a useful debugging repr of result.
// func (r *result[T]) String() string {
// 	keysbuf := getBuf()
// 	keysbuf.B = append(keysbuf.B, '[')
// 	for i := range r.keys {
// 		keysbuf.B = strconv.AppendQuote(keysbuf.B, r.keys[i].key)
// 		keysbuf.B = append(keysbuf.B, ',')
// 	}
// 	if len(keysbuf.B) > 0 {
// 		keysbuf.B = keysbuf.B[:len(keysbuf.B)-1]
// 	}
// 	keysbuf.B = append(keysbuf.B, ']')
// 	str := fmt.Sprintf("{value=%v err=%v keys=%s}", r.value, r.err, keysbuf.B)
// 	putBuf(keysbuf)
// 	return str
// }

// String returns a useful debugging repr of index.
// func (i *Index[T]) String() string {
// 	databuf := getBuf()
// 	for key, values := range i.data {
// 		databuf.WriteString("key")
// 		databuf.B = strconv.AppendQuote(databuf.B, key)
// 		databuf.B = append(databuf.B, '=')
// 		fmt.Fprintf(databuf, "%v", values)
// 		databuf.B = append(databuf.B, ' ')
// 	}
// 	if len(i.data) > 0 {
// 		databuf.B = databuf.B[:len(databuf.B)-1]
// 	}
// 	str := fmt.Sprintf("{name=%s data={%s}}", i.name, databuf.B)
// 	putBuf(databuf)
// 	return str
// }

// String returns a useful debugging repr of indexkey.
// func (i *indexkey[T]) String() string {
// 	return i.index.name + "[" + strconv.Quote(i.key) + "]"
// }
