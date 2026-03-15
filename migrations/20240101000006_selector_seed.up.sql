-- 7 fields × 3 platforms = 21 rows
-- Fields: review_text, rating, author_name, author_id, reviewed_at, product_id, xhr_pattern

INSERT INTO selector_configs (platform, field, rules) VALUES
-- Shopee
('shopee','review_text','[{"type":"css","value":".shopee-product-rating__main-text"},{"type":"xpath","value":"//div[contains(@class,''product-rating__main'')]//p"},{"type":"css","value":"[data-sqe=''review-item-content'']"}]'),
('shopee','rating','[{"type":"css","value":".shopee-product-rating__rating-star--active"},{"type":"xpath","value":"//div[@class=''shopee-star-group'']/@data-score"}]'),
('shopee','author_name','[{"type":"css","value":".shopee-product-rating__author-name"},{"type":"xpath","value":"//div[contains(@class,''author-name'')]"}]'),
('shopee','author_id','[{"type":"jsonpath","value":"$.data.ratings[*].author_username"}]'),
('shopee','reviewed_at','[{"type":"css","value":".shopee-product-rating__time"},{"type":"jsonpath","value":"$.data.ratings[*].ctime"}]'),
('shopee','product_id','[{"type":"jsonpath","value":"$.data.item_id"}]'),
('shopee','xhr_pattern','[{"type":"regex","value":"shopee\\.co\\.id/api/v[0-9]+/item/get_ratings"}]'),
-- Tokopedia
('tokopedia','review_text','[{"type":"css","value":"[data-testid=''review-description''] p"},{"type":"xpath","value":"//p[@itemprop=''description'']"},{"type":"css","value":".css-901oao"}]'),
('tokopedia','rating','[{"type":"css","value":"[data-testid=''icnStarRating'']"},{"type":"jsonpath","value":"$.data.productRevGetProductReviewList.data[*].rating"}]'),
('tokopedia','author_name','[{"type":"css","value":"[data-testid=''review-user-name'']"},{"type":"jsonpath","value":"$.data.productRevGetProductReviewList.data[*].reviewer.fullName"}]'),
('tokopedia','author_id','[{"type":"jsonpath","value":"$.data.productRevGetProductReviewList.data[*].reviewer.userID"}]'),
('tokopedia','reviewed_at','[{"type":"css","value":"[data-testid=''review-create-date'']"},{"type":"jsonpath","value":"$.data.productRevGetProductReviewList.data[*].createTime"}]'),
('tokopedia','product_id','[{"type":"jsonpath","value":"$.data.productRevGetProductReviewList.data[*].productID"}]'),
('tokopedia','xhr_pattern','[{"type":"regex","value":"tokopedia\\.com/graphql.*ProductRevGetProductReviewList"}]'),
-- Blibli
('blibli','review_text','[{"type":"css","value":".pdp-review__desc"},{"type":"xpath","value":"//div[@class=''review-content'']//p"},{"type":"jsonpath","value":"$.data[*].review"}]'),
('blibli','rating','[{"type":"css","value":".pdp-review__rating-star--active"},{"type":"jsonpath","value":"$.data[*].rating"}]'),
('blibli','author_name','[{"type":"css","value":".pdp-review__user"},{"type":"jsonpath","value":"$.data[*].name"}]'),
('blibli','author_id','[{"type":"jsonpath","value":"$.data[*].customerId"}]'),
('blibli','reviewed_at','[{"type":"css","value":".pdp-review__date"},{"type":"jsonpath","value":"$.data[*].reviewDate"}]'),
('blibli','product_id','[{"type":"jsonpath","value":"$.data[*].productSku"}]'),
('blibli','xhr_pattern','[{"type":"regex","value":"blibli\\.com/api/reviews/products/"}]')
ON CONFLICT (platform, field) DO UPDATE SET rules = EXCLUDED.rules, updated_at = NOW();
