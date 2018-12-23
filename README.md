This is a golang rewrite of Isamuni, which is a website showcasing communities, companies and professional in Siciliy (Italy).

Currently hosted [at isamuni.org](https://www.isamuni.org). Forks and contributions welcome.

[See the readme of the previous version of the project for more info](https://github.com/isamuni/isamuni)

### Design decisions

- Built as a wiki from the ground-up
- Supports both facebook and password authentication
- Users edit free-form markdown (but templates are provided)
- All pages are indexed, taking in account their markdown structure (i.e., sections and headings), to provide (full text) search. 
- Built with Go instead of Rails, to make it easier for others to contribute
- Only on-file databases are used (sqlite and bleve)