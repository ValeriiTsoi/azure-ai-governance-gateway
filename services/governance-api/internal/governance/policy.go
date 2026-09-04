package governance

func EvaluatePolicy(dataClassification string) PolicyDecision {
	const policyName = "baseline-data-classification-v1"

	switch dataClassification {
	case "restricted":
		return PolicyDecision{
			PolicyName: policyName,
			Decision:   "deny",
			Reason:     "restricted data is not permitted for model invocation by the baseline policy",
		}

	case "confidential":
		return PolicyDecision{
			PolicyName: policyName,
			Decision:   "review",
			Reason:     "confidential data requires governance review before model invocation",
		}

	default:
		return PolicyDecision{
			PolicyName: policyName,
			Decision:   "allow",
			Reason:     "data classification is permitted by the baseline policy",
		}
	}
}
