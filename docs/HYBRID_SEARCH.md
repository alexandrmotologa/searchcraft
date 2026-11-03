# Hybrid Search Architecture: BM25, Vectors, and Reciprocal Rank Fusion

SearchCraft implements hybrid retrieval to balance exact keyword matching with dense semantic recall. This document explains the underlying scoring mechanics and the rank fusion process.

---

## 1. Why Hybrid Retrieval

Full-text search engines that rely solely on BM25 struggle with vocabulary mismatch. If a user queries for "portable computer", a document titled "Lightweight Ultrabook" might receive zero keyword hits despite matching the semantic intent.

Conversely, dense vector retrieval alone suffers when users look for exact identifiers, part numbers, abbreviations, or specific product names (for example, "MX Master 3S" or "RFC 7540").

SearchCraft runs both retrievers in parallel and merges the candidate rankings using Reciprocal Rank Fusion (RRF).

---

## 2. BM25 Keyword Scoring

SearchCraft uses SQLite FTS5 for full-text indexing, tokenized via Porter stemming and Unicode character normalization.

The BM25 score for document $D$ and query $Q = \{q_1, q_2, \dots, q_n\}$ is defined as:

$$\text{Score}_{\text{BM25}}(D, Q) = \sum_{i=1}^{n} \text{IDF}(q_i) \cdot \frac{f(q_i, D) \cdot (k_1 + 1)}{f(q_i, D) + k_1 \cdot \left(1 - b + b \cdot \frac{|D|}{\text{avgdl}}\right)}$$

Where:
- $f(q_i, D)$ is the term frequency of token $q_i$ within document $D$.
- $|D|$ is the length of document $D$ in tokens, and $\text{avgdl}$ is the average document length in the corpus.
- $k_1$ controls term frequency saturation (default: $1.2$).
- $b$ controls document length normalization (default: $0.75$).

### Column Weight Boosting
SearchCraft allows assigning specific weights to document attributes. For example, matching a term in the `title` attribute can be given three times the importance of a match in the `content` body:
```sql
SELECT rowid, bm25(sc_fts, 3.0, 1.0, 1.5) FROM sc_fts WHERE sc_fts MATCH '...'
```

---

## 3. Dense Vector Cosine Similarity

Each document is converted into an embedding vector $\mathbf{v}_d \in \mathbb{R}^N$. At query time, the search term is converted into vector $\mathbf{q} \in \mathbb{R}^N$.

The semantic similarity is computed via cosine similarity:

$$\text{Sim}(\mathbf{q}, \mathbf{v}_d) = \frac{\mathbf{q} \cdot \mathbf{v}_d}{\|\mathbf{q}\| \|\mathbf{v}_d\|} = \frac{\sum_{i=1}^{N} q_i \cdot v_{d,i}}{\sqrt{\sum_{i=1}^{N} q_i^2} \cdot \sqrt{\sum_{i=1}^{N} v_{d,i}^2}}$$

SearchCraft normalizes all vectors to unit length ($\|\mathbf{v}\| = 1$) upon insertion. This simplifies cosine similarity to a dot product calculation:

$$\text{Sim}(\mathbf{q}, \mathbf{v}_d) = \sum_{i=1}^{N} q_i \cdot v_{d,i}$$

---

## 4. Reciprocal Rank Fusion (RRF)

Standard score interpolation between BM25 and vector similarity often fails because the score distributions have vastly different scales and variances.

Reciprocal Rank Fusion bypasses score distribution disparities by operating strictly on the ordinal rank positions of documents in each retrieval list.

Given two ordered lists of results:
1. $R_{\text{BM25}}$: Documents ordered by BM25 relevance.
2. $R_{\text{Vector}}$: Documents ordered by cosine similarity.

The RRF score for document $d$ is computed as:

$$\text{RRF}(d) = \frac{1}{k + \text{Rank}_{\text{BM25}}(d)} + \frac{1}{k + \text{Rank}_{\text{Vector}}(d)}$$

Where:
- $\text{Rank}(d) \in \{1, 2, 3, \dots\}$ represents the 1-based index position of document $d$.
- If a document does not appear in a candidate list, its rank in that list is treated as $\infty$, resulting in a zero contribution from that term.
- $k$ is a smoothing constant that penalizes low-ranking documents without letting top ranks dominate entirely. The standard value in research literature and SearchCraft is $k = 60$.

### Weighted Linear Blend (Hybrid Alpha)
In addition to pure ordinal RRF, SearchCraft provides a normalized linear blending parameter $\alpha \in [0, 1]$:

$$\text{Score}_{\text{Hybrid}}(d) = \alpha \cdot \text{Norm}(\text{Score}_{\text{BM25}}(d)) + (1 - \alpha) \cdot \text{Score}_{\text{Vector}}(d)$$

When $\alpha = 1.0$, search is purely keyword-driven. When $\alpha = 0.0$, search is purely vector-driven. The default setting is $\alpha = 0.5$.

---

## 5. Damerau-Levenshtein Typo Tolerance

When a search query yields fewer results than the requested threshold, the query processor evaluates token variants based on edit distance:
- Insertions: `"serch"` $\to$ `"search"`
- Deletions: `"seearch"` $\to$ `"search"`
- Substitutions: `"seyrch"` $\to$ `"search"`
- Transpositions: `"saerch"` $\to$ `"search"`

Threshold rules:
- Tokens with 1 to 3 characters: exact matches only.
- Tokens with 4 to 7 characters: 1 edit permitted.
- Tokens with 8 or more characters: up to 2 edits permitted.
