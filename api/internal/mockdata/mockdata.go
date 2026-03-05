package mockdata

import (
	"fmt"
	"math/rand"
	"time"
)

type MockObject struct {
	ID           string                 `json:"id"`
	ObjectTypeID string                 `json:"objectTypeId"`
	Properties   map[string]interface{} `json:"properties"`
}

type MockObjectType struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Properties []string `json:"properties"`
}

var ObjectTypes = []MockObjectType{
	{ID: "ot-customers", Name: "Customer", Properties: []string{"id", "name", "email", "region", "signup_date", "lifetime_value", "status"}},
	{ID: "ot-orders", Name: "Order", Properties: []string{"id", "customer_id", "product_id", "amount", "quantity", "order_date", "status", "region"}},
	{ID: "ot-products", Name: "Product", Properties: []string{"id", "name", "category", "price", "stock", "created_date"}},
}

var (
	Customers []MockObject
	Orders    []MockObject
	Products  []MockObject
)

func init() {
	r := rand.New(rand.NewSource(42))

	firstNames := []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry", "Iris", "Jack",
		"Karen", "Leo", "Mia", "Noah", "Olivia", "Paul", "Quinn", "Rose", "Sam", "Tina",
		"Uma", "Victor", "Wendy", "Xander", "Yara", "Zane", "Anna", "Ben", "Clara", "David",
		"Emma", "Felix", "Gina", "Hugo", "Ivy", "James", "Kate", "Liam", "Maya", "Nick",
		"Ora", "Pete", "Rita", "Steve", "Tara", "Uri", "Val", "Will", "Xia", "Yuri"}
	regions := []string{"North America", "Europe", "Asia Pacific", "Latin America", "Middle East"}
	statuses := []string{"active", "inactive", "churned"}
	categories := []string{"Electronics", "Clothing", "Food", "Software", "Hardware", "Services"}
	productNames := []string{"Widget Pro", "Gadget X", "Sensor V2", "Module Alpha", "Unit Beta",
		"Panel Gamma", "Board Delta", "Chip Epsilon", "Drive Zeta", "Core Eta",
		"Frame Theta", "Link Iota", "Node Kappa", "Hub Lambda", "Port Mu",
		"Cable Nu", "Plate Xi", "Dome Omicron", "Ring Pi", "Coil Rho",
		"Valve Sigma", "Pump Tau", "Motor Upsilon", "Gear Phi", "Spring Chi",
		"Bolt Psi", "Nut Omega", "Pipe Alpha2", "Tube Beta2", "Rod Gamma2",
		"Bar Delta2", "Sheet Epsilon2", "Block Zeta2", "Disk Eta2", "Sphere Theta2",
		"Cube Iota2", "Prism Kappa2", "Cone Lambda2", "Helix Mu2", "Arc Nu2",
		"Lens Xi2", "Filter Omicron2", "Seal Pi2", "Clamp Rho2", "Hinge Sigma2",
		"Joint Tau2", "Shaft Upsilon2", "Cam Phi2", "Lever Chi2", "Pulley Psi2"}
	orderStatuses := []string{"completed", "pending", "shipped", "cancelled", "refunded"}

	baseDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	// 50 Customers
	for i := 0; i < 50; i++ {
		signupDate := baseDate.AddDate(0, 0, r.Intn(730))
		Customers = append(Customers, MockObject{
			ID:           fmt.Sprintf("c-%d", i+1),
			ObjectTypeID: "ot-customers",
			Properties: map[string]interface{}{
				"id":             i + 1,
				"name":           firstNames[i],
				"email":          fmt.Sprintf("%s@example.com", firstNames[i]),
				"region":         regions[r.Intn(len(regions))],
				"signup_date":    signupDate.Format(time.RFC3339),
				"lifetime_value": float64(r.Intn(50000)) + 100,
				"status":         statuses[r.Intn(len(statuses))],
			},
		})
	}

	// 50 Products
	for i := 0; i < 50; i++ {
		Products = append(Products, MockObject{
			ID:           fmt.Sprintf("p-%d", i+1),
			ObjectTypeID: "ot-products",
			Properties: map[string]interface{}{
				"id":           i + 1,
				"name":         productNames[i],
				"category":     categories[r.Intn(len(categories))],
				"price":        float64(r.Intn(99900)+100) / 100,
				"stock":        r.Intn(1000),
				"created_date": baseDate.AddDate(0, 0, r.Intn(365)).Format(time.RFC3339),
			},
		})
	}

	// 50 Orders
	for i := 0; i < 50; i++ {
		custIdx := r.Intn(50)
		prodIdx := r.Intn(50)
		qty := r.Intn(10) + 1
		price := Products[prodIdx].Properties["price"].(float64)
		orderDate := baseDate.AddDate(0, 0, r.Intn(730))
		Orders = append(Orders, MockObject{
			ID:           fmt.Sprintf("o-%d", i+1),
			ObjectTypeID: "ot-orders",
			Properties: map[string]interface{}{
				"id":          i + 1,
				"customer_id": custIdx + 1,
				"product_id":  prodIdx + 1,
				"amount":      price * float64(qty),
				"quantity":    qty,
				"order_date":  orderDate.Format(time.RFC3339),
				"status":      orderStatuses[r.Intn(len(orderStatuses))],
				"region":      Customers[custIdx].Properties["region"],
			},
		})
	}
}

func GetObjectsByType(objectTypeID string) []MockObject {
	switch objectTypeID {
	case "ot-customers":
		return Customers
	case "ot-orders":
		return Orders
	case "ot-products":
		return Products
	}
	return nil
}

func GetLinkedObjects(objectID, objectTypeID string) []MockObject {
	switch objectTypeID {
	case "ot-customers":
		// Return orders for this customer
		custID := 0
		for _, c := range Customers {
			if c.ID == objectID {
				custID = c.Properties["id"].(int)
				break
			}
		}
		var linked []MockObject
		for _, o := range Orders {
			if int(o.Properties["customer_id"].(int)) == custID {
				linked = append(linked, o)
			}
		}
		return linked
	case "ot-orders":
		// Return customer and product for this order
		for _, o := range Orders {
			if o.ID == objectID {
				custID := int(o.Properties["customer_id"].(int))
				prodID := int(o.Properties["product_id"].(int))
				var linked []MockObject
				if custID > 0 && custID <= len(Customers) {
					linked = append(linked, Customers[custID-1])
				}
				if prodID > 0 && prodID <= len(Products) {
					linked = append(linked, Products[prodID-1])
				}
				return linked
			}
		}
	case "ot-products":
		// Return orders for this product
		prodID := 0
		for _, p := range Products {
			if p.ID == objectID {
				prodID = p.Properties["id"].(int)
				break
			}
		}
		var linked []MockObject
		for _, o := range Orders {
			if int(o.Properties["product_id"].(int)) == prodID {
				linked = append(linked, o)
			}
		}
		return linked
	}
	return nil
}

func FilterObjects(objects []MockObject, filters []Filter) []MockObject {
	if len(filters) == 0 {
		return objects
	}
	var result []MockObject
	for _, obj := range objects {
		match := true
		for _, f := range filters {
			val, ok := obj.Properties[f.Field]
			if !ok {
				match = false
				break
			}
			if !matchFilter(val, f.Operator, f.Value) {
				match = false
				break
			}
		}
		if match {
			result = append(result, obj)
		}
	}
	return result
}

type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

func matchFilter(val interface{}, op string, target interface{}) bool {
	valStr := fmt.Sprintf("%v", val)
	targetStr := fmt.Sprintf("%v", target)

	switch op {
	case "equals", "eq", "=":
		return valStr == targetStr
	case "not_equals", "neq", "!=":
		return valStr != targetStr
	case "contains":
		return len(valStr) > 0 && len(targetStr) > 0 && contains(valStr, targetStr)
	case "gt", ">":
		return toFloat(val) > toFloat(target)
	case "gte", ">=":
		return toFloat(val) >= toFloat(target)
	case "lt", "<":
		return toFloat(val) < toFloat(target)
	case "lte", "<=":
		return toFloat(val) <= toFloat(target)
	default:
		return valStr == targetStr
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case string:
		var f float64
		fmt.Sscanf(n, "%f", &f)
		return f
	}
	return 0
}

func Aggregate(objects []MockObject, field string, aggType string) float64 {
	if len(objects) == 0 {
		return 0
	}
	switch aggType {
	case "count", "value_count":
		return float64(len(objects))
	case "sum":
		var s float64
		for _, o := range objects {
			s += toFloat(o.Properties[field])
		}
		return s
	case "avg":
		var s float64
		for _, o := range objects {
			s += toFloat(o.Properties[field])
		}
		return s / float64(len(objects))
	case "min":
		m := toFloat(objects[0].Properties[field])
		for _, o := range objects[1:] {
			v := toFloat(o.Properties[field])
			if v < m {
				m = v
			}
		}
		return m
	case "max":
		m := toFloat(objects[0].Properties[field])
		for _, o := range objects[1:] {
			v := toFloat(o.Properties[field])
			if v > m {
				m = v
			}
		}
		return m
	}
	return 0
}

func GroupBy(objects []MockObject, groupField string) map[string][]MockObject {
	groups := make(map[string][]MockObject)
	for _, o := range objects {
		key := fmt.Sprintf("%v", o.Properties[groupField])
		groups[key] = append(groups[key], o)
	}
	return groups
}
