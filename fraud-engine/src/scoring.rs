use std::collections::HashMap;
use std::time::{Duration, Instant};

pub struct FraudEngine {
    buyer_history: HashMap<i64, Vec<Instant>>,
    blocked: HashMap<i64, bool>,
}

pub struct TxContext {
    pub buyer_id: i64,
    pub amount_cents: i64,
    pub new_account: bool,
    pub billing_country: String,
    pub ip_country: String,
}

#[derive(Debug, PartialEq)]
pub enum Decision {
    Approve,
    Review,
    Decline,
}

pub struct FraudResult {
    pub score: u32,
    pub decision: Decision,
    pub reasons: Vec<String>,
}

impl FraudEngine {
    pub fn new() -> Self {
        FraudEngine { buyer_history: HashMap::new(), blocked: HashMap::new() }
    }

    pub fn block(&mut self, buyer_id: i64) {
        self.blocked.insert(buyer_id, true);
    }

    pub fn unblock(&mut self, buyer_id: i64) {
        self.blocked.remove(&buyer_id);
    }

    pub fn prune(&mut self) {
        let now = Instant::now();
        let window = Duration::from_secs(600);
        self.buyer_history.retain(|_, hist| {
            hist.retain(|t| now.duration_since(*t) < window);
            !hist.is_empty()
        });
    }

    pub fn score(&mut self, tx: &TxContext) -> FraudResult {
        let mut score: i32 = 0;
        let mut reasons = Vec::new();
        if *self.blocked.get(&tx.buyer_id).unwrap_or(&false) {
            return FraudResult { score: 100, decision: Decision::Decline, reasons: vec!["buyer is blocked".into()] };
        }
        let now = Instant::now();
        let window = Duration::from_secs(600);
        let hist = self.buyer_history.entry(tx.buyer_id).or_insert_with(Vec::new);
        hist.retain(|t| now.duration_since(*t) < window);
        let recent = hist.len();
        if recent >= 5 {
            score += 45;
            reasons.push(format!("high velocity: {} tx in 10min", recent));
        } else if recent >= 3 {
            score += 25;
            reasons.push(format!("elevated velocity: {} tx in 10min", recent));
        }
        hist.push(now);
        if hist.len() > 1000 {
            let excess = hist.len() - 1000;
            hist.drain(0..excess);
        }
        if !tx.billing_country.is_empty()
            && !tx.ip_country.is_empty()
            && tx.billing_country != tx.ip_country
        {
            score += 30;
            reasons.push(format!("geo mismatch: billing {} vs ip {}", tx.billing_country, tx.ip_country));
        }
        if tx.new_account && tx.amount_cents > 50000 {
            score += 25;
            reasons.push("new account with large amount".into());
        } else if tx.new_account {
            score += 10;
            reasons.push("new account".into());
        }
        if tx.amount_cents > 200000 {
            score += 20;
            reasons.push("very large amount".into());
        }

        let score = score.clamp(0, 100) as u32;
        let decision = if score >= 70 {
            Decision::Decline
        } else if score >= 40 {
            Decision::Review
        } else {
            Decision::Approve
        };
        FraudResult { score, decision, reasons }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn clean_tx_approves() {
        let mut e = FraudEngine::new();
        let r = e.score(&TxContext { buyer_id: 1, amount_cents: 5000, new_account: false,
            billing_country: "TN".into(), ip_country: "TN".into() });
        assert_eq!(r.decision, Decision::Approve);
    }

    #[test]
    fn geo_mismatch_and_new_big_flags_review_or_decline() {
        let mut e = FraudEngine::new();
        let r = e.score(&TxContext { buyer_id: 2, amount_cents: 90000, new_account: true,
            billing_country: "TN".into(), ip_country: "RU".into() });
        assert!(r.score >= 40, "score {} should trigger review", r.score);
        assert_ne!(r.decision, Decision::Approve);
    }

    #[test]
    fn velocity_triggers_after_rapid_fire() {
        let mut e = FraudEngine::new();
        let mk = || TxContext { buyer_id: 3, amount_cents: 1000, new_account: false,
            billing_country: "TN".into(), ip_country: "TN".into() };
        for _ in 0..5 { e.score(&mk()); }
        let r = e.score(&mk());
        assert!(r.reasons.iter().any(|s| s.contains("velocity")), "velocity should flag: {:?}", r.reasons);
    }

    #[test]
    fn blocked_buyer_declined() {
        let mut e = FraudEngine::new();
        e.block(99);
        let r = e.score(&TxContext { buyer_id: 99, amount_cents: 100, new_account: false,
            billing_country: "TN".into(), ip_country: "TN".into() });
        assert_eq!(r.decision, Decision::Decline);
        assert_eq!(r.score, 100);
    }
}
