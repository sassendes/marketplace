use std::collections::HashMap;

#[derive(Clone)]
pub struct Document {
    pub id: i64,
    pub title: String,
    pub body: String,
}

struct Posting {
    doc_id: i64,
    term_freq: u32,
}

#[derive(Default)]
pub struct SearchIndex {
    postings: HashMap<String, Vec<Posting>>,
    doc_terms: HashMap<i64, Vec<String>>,
    doc_len: HashMap<i64, u32>,
    docs: HashMap<i64, Document>,
    total_len: u64,
}

const MAX_QUERY_TERMS: usize = 32;

impl SearchIndex {
    pub fn new() -> Self {
        Self::default()
    }

    fn tokenize(text: &str) -> Vec<String> {
        text.to_lowercase()
            .split(|c: char| !c.is_alphanumeric())
            .filter(|s| !s.is_empty() && s.len() <= 64)
            .map(|s| s.to_string())
            .collect()
    }

    pub fn add(&mut self, doc: Document) {
        self.remove(doc.id);
        let combined = format!("{} {} {}", doc.title, doc.title, doc.body);
        let tokens = Self::tokenize(&combined);
        let len = tokens.len() as u32;
        let mut tf: HashMap<String, u32> = HashMap::new();
        for t in tokens {
            *tf.entry(t).or_insert(0) += 1;
        }
        let mut terms = Vec::with_capacity(tf.len());
        for (term, freq) in tf {
            self.postings
                .entry(term.clone())
                .or_default()
                .push(Posting { doc_id: doc.id, term_freq: freq });
            terms.push(term);
        }
        self.doc_terms.insert(doc.id, terms);
        self.doc_len.insert(doc.id, len);
        self.total_len += len as u64;
        self.docs.insert(doc.id, doc);
    }

    pub fn remove(&mut self, id: i64) -> bool {
        let Some(terms) = self.doc_terms.remove(&id) else {
            return false;
        };
        for term in terms {
            if let Some(list) = self.postings.get_mut(&term) {
                list.retain(|p| p.doc_id != id);
                if list.is_empty() {
                    self.postings.remove(&term);
                }
            }
        }
        if let Some(len) = self.doc_len.remove(&id) {
            self.total_len -= len as u64;
        }
        self.docs.remove(&id);
        true
    }

    fn avg_doc_len(&self) -> f64 {
        if self.docs.is_empty() {
            0.0
        } else {
            self.total_len as f64 / self.docs.len() as f64
        }
    }

    pub fn search(&self, query: &str, limit: usize) -> Vec<(Document, f64)> {
        const K1: f64 = 1.2;
        const B: f64 = 0.75;
        let mut terms = Self::tokenize(query);
        terms.sort();
        terms.dedup();
        terms.truncate(MAX_QUERY_TERMS);
        if terms.is_empty() || self.docs.is_empty() {
            return vec![];
        }
        let n = self.docs.len() as f64;
        let avg = self.avg_doc_len().max(1.0);
        let mut scores: HashMap<i64, f64> = HashMap::new();
        for term in &terms {
            let candidates: Vec<&String> = if term.len() >= 3 {
                self.postings.keys().filter(|k| *k == term || k.starts_with(term.as_str())).collect()
            } else {
                self.postings.get_key_value(term).map(|(k, _)| vec![k]).unwrap_or_default()
            };
            for key in candidates {
                let postings = &self.postings[key];
                let weight = if key == term { 1.0 } else { 0.6 };
                let df = postings.len() as f64;
                let idf = ((n - df + 0.5) / (df + 0.5) + 1.0).ln();
                for p in postings {
                    let dl = *self.doc_len.get(&p.doc_id).unwrap_or(&0) as f64;
                    let tf = p.term_freq as f64;
                    let score = idf * (tf * (K1 + 1.0)) / (tf + K1 * (1.0 - B + B * (dl / avg)));
                    *scores.entry(p.doc_id).or_insert(0.0) += score * weight;
                }
            }
        }
        let mut ranked: Vec<(Document, f64)> = scores
            .into_iter()
            .filter_map(|(id, s)| self.docs.get(&id).map(|d| (d.clone(), s)))
            .collect();
        ranked.sort_by(|a, b| b.1.partial_cmp(&a.1).unwrap_or(std::cmp::Ordering::Equal).then(a.0.id.cmp(&b.0.id)));
        ranked.truncate(limit);
        ranked
    }

    pub fn len(&self) -> usize {
        self.docs.len()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn doc(id: i64, title: &str, body: &str) -> Document {
        Document { id, title: title.into(), body: body.into() }
    }

    #[test]
    fn bm25_ranks_relevant_first() {
        let mut idx = SearchIndex::new();
        idx.add(doc(1, "Red running shoes", "comfortable shoes for running and sport"));
        idx.add(doc(2, "Blue jeans", "denim trousers casual wear"));
        idx.add(doc(3, "Running socks", "socks for runners athletic"));
        let hits = idx.search("running shoes", 10);
        assert_eq!(hits[0].0.id, 1);
        assert!(!hits.iter().any(|(d, _)| d.id == 2));
    }

    #[test]
    fn empty_query_returns_nothing() {
        let mut idx = SearchIndex::new();
        idx.add(doc(1, "x", "y"));
        assert!(idx.search("", 10).is_empty());
        assert!(idx.search("nonexistentterm", 10).is_empty());
    }

    #[test]
    fn re_adding_a_document_does_not_duplicate_postings() {
        let mut idx = SearchIndex::new();
        idx.add(doc(1, "kettle", "gooseneck kettle"));
        idx.add(doc(1, "kettle", "gooseneck kettle"));
        idx.add(doc(2, "grinder", "burr grinder"));
        assert_eq!(idx.len(), 2);
        assert_eq!(idx.postings["kettle"].len(), 1);
    }

    #[test]
    fn remove_drops_document_from_results() {
        let mut idx = SearchIndex::new();
        idx.add(doc(1, "tent", "two person tent"));
        assert!(idx.remove(1));
        assert!(idx.search("tent", 10).is_empty());
        assert_eq!(idx.len(), 0);
        assert!(idx.postings.is_empty());
    }

    #[test]
    fn prefix_matches_partial_words() {
        let mut idx = SearchIndex::new();
        idx.add(doc(1, "Headphones", "wireless"));
        idx.add(doc(2, "Keyboard", "mechanical"));
        let hits = idx.search("headph", 10);
        assert_eq!(hits.len(), 1);
        assert_eq!(hits[0].0.id, 1);
    }
}
