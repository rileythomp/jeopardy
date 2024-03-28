select id, email,
COALESCE(NULLIF(raw_user_meta_data->>'display_name', ''), raw_user_meta_data->>'full_name') as display_name,
COALESCE(NULLIF(raw_user_meta_data->>'user_img_url', ''), raw_user_meta_data->>'avatar_url') as img_url,
confirmed_at
from auth.users
where raw_user_meta_data->>'display_name' = $1
or raw_user_meta_data->>'full_name' = $1;