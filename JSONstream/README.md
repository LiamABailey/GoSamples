#### Parsing JSON streams

At work, I needed to parse some large JSON files stored in AWS S3 in order to
insert some of the data into a Redshift database for later use. I had two
document formats, and wanted something that could read documents of
arbitrary size.

I settled on a solution similar to the one shared here, but in Python (most
of the work my team does is in Python, and I try to stick to that to make
it easier for other team members to read/use/learn from/etc. my code). That
solution used processing and loading functions passed to a managing function that
applied them on blocks of data from a file steam.

I wanted to see if I could come up with a reasonable solution for the same
problem in Go, using interfaces - the result is shared here.

There are a number of possible improvements:
-   Error handling for struct/JSON data-format mismatch
-   Attempt to resolve structural issues in the JSON (missing brackets, commas)
-   better handling of multi-byte characters
-   attempt to parse out multiple JSON records in one go (this would require major restructuring)


Note: Sample JSON generated using [JSON GENERATOR](https://www.json-generator.com/)
